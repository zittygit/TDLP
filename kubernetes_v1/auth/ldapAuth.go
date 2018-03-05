package auth

import (
	"github.com/sona-tar/ldapc"
	"gopkg.in/ldap.v2"
)

var (
	LdapServer   string
	BindDN       string
	BindPassword string
	UserDN       string
	GroupDN      string
)

func LdapAuthenticateUser(username string, password string) *ldap.Entry {
	ldapclient := &ldapc.Client{
		Protocol:  ldapc.LDAP,
		Host:      LdapServer,
		Port:      389,
		TLSConfig: nil,
		Bind: &ldapc.DirectBind{
			UserDN: UserDN,
			Filter: "(&(objectClass=posixAccount)(uid=%s))",
		},
	}
	entry, _ := ldapclient.Authenticate(username, password)
	return entry
}
