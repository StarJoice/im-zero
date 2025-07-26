package logic

import (
	"context"
	"strings"

	"im-zero/app/usercenter/cmd/rpc/internal/svc"
	"im-zero/app/usercenter/cmd/rpc/pb"
	"im-zero/pkg/xerrs"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

type SearchUsersByNicknameLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchUsersByNicknameLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchUsersByNicknameLogic {
	return &SearchUsersByNicknameLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 根据昵称搜索用户
func (l *SearchUsersByNicknameLogic) SearchUsersByNickname(in *pb.SearchUsersByNicknameReq) (*pb.SearchUsersByNicknameResp, error) {
	// 参数验证
	keyword := strings.TrimSpace(in.Keyword)
	if len(keyword) < 2 {
		return nil, errors.Wrapf(xerrs.NewErrCodeMsg(xerrs.PARAM_ERROR, "keyword too short"), "keyword=%s", keyword)
	}
	if in.Page <= 0 {
		in.Page = 1
	}
	if in.Limit <= 0 || in.Limit > 50 {
		in.Limit = 10
	}

	// 搜索用户
	users, err := l.svcCtx.UserModel.SearchByNickname(l.ctx, keyword, in.Page, in.Limit)
	if err != nil {
		return nil, errors.Wrapf(err, "search users by nickname failed")
	}

	// 获取总数
	total, err := l.svcCtx.UserModel.CountByNickname(l.ctx, keyword)
	if err != nil {
		return nil, errors.Wrapf(err, "count users by nickname failed")
	}

	// 转换结果
	var pbUsers []*pb.User
	for _, user := range users {
		pbUsers = append(pbUsers, &pb.User{
			Id:       user.Id,
			Mobile:   user.Mobile,
			Nickname: user.Nickname,
			Avatar:   user.Avatar,
			Sign:     user.Sign,
		})
	}

	return &pb.SearchUsersByNicknameResp{
		Users: pbUsers,
		Total: int32(total),
	}, nil
}
