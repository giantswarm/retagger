package main

import (
	"gopkg.in/check.v1"
)

const blockedRegistriesConf = "./fixtures/blocked-registries.conf"
const blockedErrorRegex = `.*registry registry-blocked.com is blocked in .*`

func (s *SkopeoSuite) TestCopyBlockedSource(c *check.C) {
	assertSkopeoFails(c, blockedErrorRegex,
		"--registries-conf", blockedRegistriesConf, "copy",
		"docker://registry-blocked.com/image:test",
		"docker://registry-unblocked.com/image:test")
}

func (s *SkopeoSuite) TestCopyBlockedDestination(c *check.C) {
	assertSkopeoFails(c, blockedErrorRegex,
		"--registries-conf", blockedRegistriesConf, "copy",
		"docker://registry-unblocked.com/image:test",
		"docker://registry-blocked.com/image:test")
}

func (s *SkopeoSuite) TestInspectBlocked(c *check.C) {
	assertSkopeoFails(c, blockedErrorRegex,
		"--registries-conf", blockedRegistriesConf, "inspect",
		"docker://registry-blocked.com/image:test")
}

func (s *SkopeoSuite) TestDeleteBlocked(c *check.C) {
	assertSkopeoFails(c, blockedErrorRegex,
		"--registries-conf", blockedRegistriesConf, "delete",
		"docker://registry-blocked.com/image:test")
}
