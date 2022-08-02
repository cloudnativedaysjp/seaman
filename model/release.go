package model

import (
	"fmt"
	"strings"
)

type OrgRepo struct {
	org  string
	repo string
}

func NewOrgRepo(str string) (OrgRepo, error) {
	s := strings.Split(str, "__")
	if len(s) != 2 {
		return OrgRepo{}, fmt.Errorf("callbackValue (%s) is not expected", str)
	}
	return OrgRepo{s[0], s[1]}, nil
}

func (m OrgRepo) Org() string {
	return m.org
}

func (m OrgRepo) Repo() string {
	return m.repo
}

func (m OrgRepo) PullRequestUrl(number int) string {
	return fmt.Sprintf("https://github.com/%s/%s/pull/%d", m.org, m.repo, number)
}

func (m OrgRepo) WithLevel(level string) OrgRepoLevel {
	return OrgRepoLevel{m, level}
}

type OrgRepoLevel struct {
	OrgRepo
	level string
}

func NewOrgRepoLevel(str string) (OrgRepoLevel, error) {
	s := strings.Split(str, "__")
	if len(s) != 3 {
		return OrgRepoLevel{}, fmt.Errorf("callbackValue (%s) is not expected", str)
	}
	return OrgRepoLevel{OrgRepo{s[0], s[1]}, s[2]}, nil
}

func (m OrgRepoLevel) String() string {
	return fmt.Sprintf("%s__%s__%s", m.org, m.repo, m.level)
}

func (m OrgRepoLevel) Level() string {
	return m.level
}
