package auth

import (
	"github.com/sonatard/go-ldapc"
	"gopkg.in/ldap.v2"
)

var (
	LDAPServer   string
	BindDN       string
	BindPassWord string
	UserDN       string
	GroupDN      string
)

func LdapAuthenticateUser(userName string, passWord string) (*ldap.Entry, error) {
	ldapclient := &ldapc.Client{
		Protocol:  ldapc.LDAP,
		Host:      LDAPServer,
		Port:      389,
		TLSConfig: nil,
		Bind: &ldapc.DirectBind{
			UserDN: UserDN,
			Filter: "(&(objectClass=posixAccount)(uid=%s))",
		},
	}
	return ldapclient.Authenticate(userName, passWord)
}
