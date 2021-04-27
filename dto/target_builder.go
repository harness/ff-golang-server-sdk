package dto

import (
	"github.com/drone/ff-golang-server-sdk/evaluation"
)

// TargetBuilderInterface used for fluent builder methods
type TargetBuilderInterface interface {
	IP(string) TargetBuilderInterface
	Country(string) TargetBuilderInterface
	Email(string) TargetBuilderInterface
	Firstname(string) TargetBuilderInterface
	Lastname(string) TargetBuilderInterface
	Name(string) TargetBuilderInterface
	Anonymous(bool) TargetBuilderInterface
	Custom(name string, value interface{}) TargetBuilderInterface
	Build() evaluation.Target
}

// TargetBuilder structure for building targets
type targetBuilder struct {
	*evaluation.Target
}

// NewTargetBuilder constructing TargetBuilder instance
func NewTargetBuilder(identifier string) TargetBuilderInterface {
	return &targetBuilder{
		Target: &evaluation.Target{
			Identifier: identifier,
		},
	}
}

// IP set Ip address to target object
func (b *targetBuilder) IP(ip string) TargetBuilderInterface {
	b.Custom("ip", ip)
	return b
}

// Country set country of target object
func (b *targetBuilder) Country(country string) TargetBuilderInterface {
	b.Custom("country", country)
	return b
}

// Email set email address to target object
func (b *targetBuilder) Email(email string) TargetBuilderInterface {
	b.Custom("email", email)
	return b
}

// Firstname set firstname of target object
func (b *targetBuilder) Firstname(firstname string) TargetBuilderInterface {
	b.Custom("first_name", firstname)
	return b
}

// Lastname set lastname of target object
func (b *targetBuilder) Lastname(lastname string) TargetBuilderInterface {
	b.Custom("last_name", lastname)
	return b
}

// Name target name object
func (b *targetBuilder) Name(name string) TargetBuilderInterface {
	b.Custom("name", name)
	return b
}

// Anonymous target object
func (b *targetBuilder) Anonymous(value bool) TargetBuilderInterface {
	b.Target.Anonymous = value
	return b
}

// Custom object
func (b *targetBuilder) Custom(key string, value interface{}) TargetBuilderInterface {
	if b.Target.Attributes == nil {
		b.Target.Attributes = make(map[string]interface{})
	}

	b.Target.Attributes[key] = value
	return b
}

// Build returns target object
func (b *targetBuilder) Build() evaluation.Target {
	return *b.Target
}
