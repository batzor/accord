package accord

import (
	pb "github.com/qvntm/accord/pb"
)

// Role is a role of a user within a channel.
type Role int

const (
	UnknownRole Role = iota
	SubscriberRole
	MemberRole
	AdminRole
	SuperadminRole
)

var AccordToPBRoles = map[Role]pb.Role{
	UnknownRole:    pb.Role_UNKNOWN_ROLE,
	SubscriberRole: pb.Role_SUBSCRIBER,
	MemberRole:     pb.Role_MEMBER,
	AdminRole:      pb.Role_ADMIN,
	SuperadminRole: pb.Role_SUPERADMIN,
}

var PBToAccordRoles = map[pb.Role]Role{
	pb.Role_UNKNOWN_ROLE: UnknownRole,
	pb.Role_SUBSCRIBER:   SubscriberRole,
	pb.Role_MEMBER:       MemberRole,
	pb.Role_ADMIN:        AdminRole,
	pb.Role_SUPERADMIN:   SuperadminRole,
}
