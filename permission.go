package accord

type Permission int

const (
	// ReadPermission is for subscribed users or any member of the channel
	ReadPermission Permission = iota
	// WritePermission includes writing, modifying, and deletion of messages
	WritePermission
	// DeletePermission allows deleting othes users' messages
	DeletePermission
	// ModifyPermission is for modifying channel configurations
	ModifyPermission
	// KickPermission is for kicking users out of the channel
	KickPermission
	// BanPermission is for banning users
	BanPermission
)

type UserPermission struct {
	permissionValue uint16
}

func NewEmptyPermission() *UserPermission {
	return &UserPermission{permissionValue: 0}
}

func (userPerm *UserPermission) Has(perm Permission) bool {
	if x := userPerm.permissionValue & (1 << perm); x != 0 {
		return true
	}
	return false
}

func (userPerm *UserPermission) Add(perm Permission) {
	if !userPerm.Has(perm) {
		userPerm.permissionValue = userPerm.permissionValue ^ (1 << perm)
	}
}

func (userPerm *UserPermission) Remove(perm Permission) {
	if userPerm.Has(perm) {
		userPerm.permissionValue = userPerm.permissionValue ^ (1 << perm)
	}
}

func (userPerm *UserPermission) AddAll() {
	userPerm.permissionValue = 1<<16 - 1
}

func (userPerm *UserPermission) RemoveAll() {
	userPerm.permissionValue = 0
}
