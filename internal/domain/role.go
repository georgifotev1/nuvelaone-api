package domain

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleOwner  Role = "owner"
	RoleMember Role = "member"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleAdmin, RoleOwner, RoleMember:
		return true
	}
	return false
}
