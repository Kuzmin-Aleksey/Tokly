package server

import (
	"FairLAP/internal/domain/service/groups"
	"FairLAP/pkg/failure"
	"net/http"
	"strconv"
)

type GroupsServer struct {
	groups *groups.Service
}

func NewGroupsServer(groups *groups.Service) *GroupsServer {
	return &GroupsServer{
		groups: groups,
	}
}

func (s *GroupsServer) CreateGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	lapId, err := strconv.Atoi(r.FormValue("lap_id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid lap_id"))
		return
	}

	id, err := s.groups.CreateGroup(ctx, lapId)
	if err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}

	writeJson(ctx, w, IdResponse{Id: id}, http.StatusOK)
}

func (s *GroupsServer) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		writeAndLogErr(ctx, w, failure.NewInvalidRequestError("invalid id"))
	}
	if err := s.groups.DeleteGroup(ctx, id); err != nil {
		writeAndLogErr(ctx, w, err)
		return
	}
}
