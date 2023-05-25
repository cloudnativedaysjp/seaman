package cosme

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/webhooks/v6/github"
	"golang.org/x/exp/slog"

	"github.com/cloudnativedaysjp/seaman/pkg/log"
	"github.com/cloudnativedaysjp/seaman/pkg/utils"
)

type issueCommentHandler func(ctx context.Context, payload github.IssueCommentPayload, args []string) error

type handler struct {
	c        chan data
	commands map[string]issueCommentHandler
	log      *slog.Logger
	hook     *github.Webhook
}

type data struct {
	payload github.IssueCommentPayload
	ctx     context.Context
}

func New(logger *slog.Logger, secret string) (*handler, error) {
	if logger == nil {
		logger = slog.Default()
	}
	hook, err := github.New(github.Options.Secret(secret))
	if err != nil {
		return nil, err
	}
	return &handler{
		make(chan data),
		make(map[string]issueCommentHandler),
		logger.With("package", "cosme"),
		hook,
	}, nil
}

func (h *handler) WithCommand(command string, handler issueCommentHandler) *handler {
	h.commands[command] = handler
	return h
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get payload
	payloadRaw, err := h.hook.Parse(r, github.IssueCommentEvent)
	if err != nil {
		return
	}
	payload, ok := payloadRaw.(github.IssueCommentPayload)
	if !ok {
		return
	}
	if payload.Action != "created" {
		return
	}

	// if the user who send the IssueComment is unauthorized, skip
	roles := []string{"OWNER", "COLLABORATOR", "CONTRIBUTOR", "MEMBER"}
	if !utils.Contains(roles, payload.Comment.AuthorAssociation) {
		h.log.Info("unauthorized the user who send the IssueComment")
		return
	}

	// hook handler

	if len(strings.Fields(payload.Comment.Body)) == 0 {
		h.log.Info("invalid command: args.length == 0")
		return
	}
	h.c <- data{payload, r.Context()}

	w.WriteHeader(http.StatusOK)
}

func (h *handler) RunBackground() {
	for d := range h.c {
		select {
		case <-d.ctx.Done():
			h.log.Info("context has already exceeded")
			continue
		default:
		}

		ctx := context.Background()
		ctx = log.IntoContext(ctx, h.log)
		commandAndArgs := strings.Fields(d.payload.Comment.Body)

		for registeredCommand, handler := range h.commands {
			if commandAndArgs[0] == registeredCommand {
				if err := handler(ctx, d.payload, commandAndArgs[1:]); err != nil {
					h.log.Error(fmt.Sprintf("internal server error: %v", err),
						log.KeyDetail, err)
					return
				}
			}
		}
	}
}
