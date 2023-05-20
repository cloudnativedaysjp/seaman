package cosme

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-playground/webhooks/v6/github"
	"golang.org/x/exp/slog"

	"github.com/cloudnativedaysjp/seaman/pkg/utils"
)

type issueCommentHandler func(ctx context.Context, payload github.IssueCommentPayload, args []string) error

type handler struct {
	logger   *slog.Logger
	hook     *github.Webhook
	commands map[string]issueCommentHandler
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
		logger.With("package", "cosme"),
		hook,
		make(map[string]issueCommentHandler),
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
		h.logger.Info("unauthorized the user who send the IssueComment")
		return
	}

	// hook handler
	commandAndArgs := strings.Fields(payload.Comment.Body)
	if len(commandAndArgs) == 0 {
		h.logger.Info("invalid command: args.length == 0")
		return
	}
	ctx := r.Context()
	fmt.Println(ctx.Deadline())
	for registeredCommand, handler := range h.commands {
		if commandAndArgs[0] == registeredCommand {
			if err := handler(r.Context(), payload, commandAndArgs[1:]); err != nil {
				h.logger.Error(fmt.Sprintf("internal server error: %v", err))
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}
	}

	w.WriteHeader(http.StatusOK)
}
