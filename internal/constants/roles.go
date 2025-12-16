package constants

type Role string

const (
	RoleMasterAdmin Role = "masteradmin"
	RoleAdmin       Role = "admin"
)

func (r Role) String() string {
	return string(r)
}
