package okta

import (
	"context"
	"fmt"

	"github.com/okta/okta-sdk-golang/okta"
	"github.com/okta/okta-sdk-golang/okta/query"
)

type OktaService struct {
	Client *okta.Client
}

type OktaConfig struct {
	Org   string
	Token string
}

func NewOktaService(oc OktaConfig) *OktaService {
	ocl, _ := okta.NewClient(
		context.TODO(),
		okta.WithOrgUrl(oc.Org),
		okta.WithToken(oc.Token),
	)
	return &OktaService{
		Client: ocl,
	}
}

func (o *OktaService) getGroup(groupName string) (*okta.Group, error) {
	q := query.NewQueryParams(query.WithSearch(fmt.Sprintf("profile.name eq \"%s\"", groupName)))
	groups, _, err := o.Client.Group.ListGroups(q)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %v", err)
	}
	if len(groups) == 1 {
		return groups[0], nil
	}
	return nil, fmt.Errorf("failed to get group")
}

func (o *OktaService) GetUsersFromGroup(groupName string) ([]*okta.User, error) {
	group, err := o.getGroup(groupName)
	if err != nil {
		return nil, err
	}
	users, _, err := o.Client.Group.ListGroupUsers(group.Id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get users from group: %v", err)
	}
	users = o.filterActiveUsers(users)
	return users, nil
}

func (o *OktaService) filterActiveUsers(users []*okta.User) []*okta.User {
	activeUsers := []*okta.User{}
	for uk, uv := range users {
		if uv.Status == "ACTIVE" {
			activeUsers = append(activeUsers, users[uk])
		}
	}
	return activeUsers
}
