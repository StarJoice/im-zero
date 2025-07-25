package logic

import (
	"context"
	"strings"

	"im-zero/app/friend/cmd/rpc/friend"
	"im-zero/app/friend/cmd/rpc/internal/svc"
	"im-zero/app/usercenter/cmd/rpc/usercenter"
	"im-zero/pkg/tool"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type SearchUsersLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchUsersLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchUsersLogic {
	return &SearchUsersLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 搜索用户
func (l *SearchUsersLogic) SearchUsers(in *friend.SearchUsersReq) (*friend.SearchUsersResp, error) {
	// 记录搜索开始
	l.Logger.Infow("Start searching users",
		logx.Field("userId", in.UserId),
		logx.Field("keyword", in.Keyword),
		logx.Field("page", in.Page),
		logx.Field("limit", in.Limit),
	)

	// 参数验证
	if in.UserId <= 0 {
		l.Logger.Errorw("Invalid user id", logx.Field("userId", in.UserId))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "invalid user id"), "userId=%d", in.UserId)
	}
	if len(strings.TrimSpace(in.Keyword)) < 2 {
		l.Logger.Errorw("Keyword too short", logx.Field("keyword", in.Keyword))
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "keyword too short"), "keyword=%s", in.Keyword)
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.Limit <= 0 || in.Limit > 20 {
		in.Limit = 10
	}

	keyword := strings.TrimSpace(in.Keyword)
	var users []*friend.UserInfo

	// 先尝试手机号精确搜索
	if tool.ValidateMobile(keyword) {
		l.Logger.Infow("Searching by mobile", logx.Field("mobile", keyword))
		
		userInfo, err := l.svcCtx.UsercenterRpc.GetUserByMobile(l.ctx, &usercenter.GetUserByMobileReq{
			Mobile: keyword,
		})
		if err == nil && userInfo.User != nil {
			// 排除自己
			if userInfo.User.Id != in.UserId {
				// 检查是否已经是好友
				isFriend, _ := l.svcCtx.ImFriendModel.CheckFriendship(l.ctx, in.UserId, userInfo.User.Id)
				
				// 检查拉黑状态
				isBlocked, hasBlocked, _ := l.svcCtx.ImUserBlacklistModel.CheckMutualBlocked(l.ctx, in.UserId, userInfo.User.Id)
				
				user := &friend.UserInfo{
					Id:       userInfo.User.Id,
					Mobile:   userInfo.User.Mobile,
					Nickname: userInfo.User.Nickname,
					Avatar:   userInfo.User.Avatar,
					Sign:     userInfo.User.Sign,
					IsFriend: isFriend,
					IsBlocked: isBlocked || hasBlocked,
				}
				users = append(users, user)
			}
		} else {
			l.Logger.Infow("User not found by mobile", logx.Field("mobile", keyword))
		}
	} else {
		// 按昵称模糊搜索
		l.Logger.Infow("Searching by nickname", logx.Field("keyword", keyword))
		
		userList, err := l.svcCtx.UsercenterRpc.SearchUsersByNickname(l.ctx, &usercenter.SearchUsersByNicknameReq{
			Keyword: keyword,
			Page:    in.Page,
			Limit:   in.Limit,
		})
		if err != nil {
			l.Logger.Errorw("Search users by nickname failed", 
				logx.Field("keyword", keyword), 
				logx.Field("error", err.Error()))
			return nil, errors.Wrapf(err, "search users by nickname failed")
		}
		
		for _, userInfo := range userList.Users {
			// 排除自己
			if userInfo.Id != in.UserId {
				// 检查是否已经是好友
				isFriend, _ := l.svcCtx.ImFriendModel.CheckFriendship(l.ctx, in.UserId, userInfo.Id)
				
				// 检查拉黑状态
				isBlocked, hasBlocked, _ := l.svcCtx.ImUserBlacklistModel.CheckMutualBlocked(l.ctx, in.UserId, userInfo.Id)
				
				user := &friend.UserInfo{
					Id:       userInfo.Id,
					Mobile:   userInfo.Mobile,
					Nickname: userInfo.Nickname,
					Avatar:   userInfo.Avatar,
					Sign:     userInfo.Sign,
					IsFriend: isFriend,
					IsBlocked: isBlocked || hasBlocked,
				}
				users = append(users, user)
			}
		}
	}

	l.Logger.Infow("Search users successfully",
		logx.Field("userId", in.UserId),
		logx.Field("keyword", keyword),
		logx.Field("resultCount", len(users)),
	)

	return &friend.SearchUsersResp{
		Users: users,
		Total: int32(len(users)),
	}, nil
}
