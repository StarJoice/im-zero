// Code generated by goctl. DO NOT EDIT.
// goctl 1.8.3
// Source: group.proto

package server

import (
	"context"

	"im-zero/app/group/cmd/rpc/group"
	"im-zero/app/group/cmd/rpc/internal/logic"
	"im-zero/app/group/cmd/rpc/internal/svc"
)

type GroupServer struct {
	svcCtx *svc.ServiceContext
	group.UnimplementedGroupServer
}

func NewGroupServer(svcCtx *svc.ServiceContext) *GroupServer {
	return &GroupServer{
		svcCtx: svcCtx,
	}
}

// 创建群组
func (s *GroupServer) CreateGroup(ctx context.Context, in *group.CreateGroupReq) (*group.CreateGroupResp, error) {
	l := logic.NewCreateGroupLogic(ctx, s.svcCtx)
	return l.CreateGroup(in)
}

// 获取群组信息
func (s *GroupServer) GetGroupInfo(ctx context.Context, in *group.GetGroupInfoReq) (*group.GetGroupInfoResp, error) {
	l := logic.NewGetGroupInfoLogic(ctx, s.svcCtx)
	return l.GetGroupInfo(in)
}

// 更新群组信息
func (s *GroupServer) UpdateGroup(ctx context.Context, in *group.UpdateGroupReq) (*group.UpdateGroupResp, error) {
	l := logic.NewUpdateGroupLogic(ctx, s.svcCtx)
	return l.UpdateGroup(in)
}

// 解散群组
func (s *GroupServer) DissolveGroup(ctx context.Context, in *group.DissolveGroupReq) (*group.DissolveGroupResp, error) {
	l := logic.NewDissolveGroupLogic(ctx, s.svcCtx)
	return l.DissolveGroup(in)
}

// 邀请用户入群
func (s *GroupServer) InviteUsers(ctx context.Context, in *group.InviteUsersReq) (*group.InviteUsersResp, error) {
	l := logic.NewInviteUsersLogic(ctx, s.svcCtx)
	return l.InviteUsers(in)
}

// 移除群成员
func (s *GroupServer) RemoveMembers(ctx context.Context, in *group.RemoveMembersReq) (*group.RemoveMembersResp, error) {
	l := logic.NewRemoveMembersLogic(ctx, s.svcCtx)
	return l.RemoveMembers(in)
}

// 退出群组
func (s *GroupServer) LeaveGroup(ctx context.Context, in *group.LeaveGroupReq) (*group.LeaveGroupResp, error) {
	l := logic.NewLeaveGroupLogic(ctx, s.svcCtx)
	return l.LeaveGroup(in)
}

// 获取群成员列表
func (s *GroupServer) GetGroupMembers(ctx context.Context, in *group.GetGroupMembersReq) (*group.GetGroupMembersResp, error) {
	l := logic.NewGetGroupMembersLogic(ctx, s.svcCtx)
	return l.GetGroupMembers(in)
}

// 设置成员角色
func (s *GroupServer) SetMemberRole(ctx context.Context, in *group.SetMemberRoleReq) (*group.SetMemberRoleResp, error) {
	l := logic.NewSetMemberRoleLogic(ctx, s.svcCtx)
	return l.SetMemberRole(in)
}

// 禁言成员
func (s *GroupServer) MuteMembers(ctx context.Context, in *group.MuteMembersReq) (*group.MuteMembersResp, error) {
	l := logic.NewMuteMembersLogic(ctx, s.svcCtx)
	return l.MuteMembers(in)
}

// 获取用户的群组列表
func (s *GroupServer) GetUserGroups(ctx context.Context, in *group.GetUserGroupsReq) (*group.GetUserGroupsResp, error) {
	l := logic.NewGetUserGroupsLogic(ctx, s.svcCtx)
	return l.GetUserGroups(in)
}

// 检查用户是否在群中
func (s *GroupServer) CheckMembership(ctx context.Context, in *group.CheckMembershipReq) (*group.CheckMembershipResp, error) {
	l := logic.NewCheckMembershipLogic(ctx, s.svcCtx)
	return l.CheckMembership(in)
}

// 发送群消息
func (s *GroupServer) SendGroupMessage(ctx context.Context, in *group.SendGroupMessageReq) (*group.SendGroupMessageResp, error) {
	l := logic.NewSendGroupMessageLogic(ctx, s.svcCtx)
	return l.SendGroupMessage(in)
}

// 获取群聊记录
func (s *GroupServer) GetGroupHistory(ctx context.Context, in *group.GetGroupHistoryReq) (*group.GetGroupHistoryResp, error) {
	l := logic.NewGetGroupHistoryLogic(ctx, s.svcCtx)
	return l.GetGroupHistory(in)
}
