package accord

import pb "github.com/qvntm/accord/pb"

// Permission represents actions allowed to the role within a channel.
type Permission int

const (
	// UnknownPermission is needed as a part of mapping to unknown message
	// from "pb" package.
	UnknownPermission Permission = iota
	// ReadPermission is for subscribed users or any member of the channel
	ReadPermission
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
	// AssignRolePermission is for assignment of roles to all channel's users
	AssignRolePermission
	// RemoveChannelPermission is a permission to permanently remove the channel
	// and all of its data.
	RemoveChannelPermission
)

// AccordToPBPermissions is a mapping from objects of "Permission" type of this
// package to the objects from "pb" package.
var AccordToPBPermissions = map[Permission]pb.Permission{
	UnknownPermission:       pb.Permission_UNKNOWN_PERMISSION,
	ReadPermission:          pb.Permission_READ,
	WritePermission:         pb.Permission_WRITE,
	DeletePermission:        pb.Permission_DELETE,
	ModifyPermission:        pb.Permission_MODIFY,
	KickPermission:          pb.Permission_KICK,
	BanPermission:           pb.Permission_BAN,
	AssignRolePermission:    pb.Permission_ASSIGN_ROLE,
	RemoveChannelPermission: pb.Permission_REMOVE_CHANNEL,
}
