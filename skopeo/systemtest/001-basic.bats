#!/usr/bin/env bats
#
# Simplest set of skopeo tests. If any of these fail, we have serious problems.
#

load helpers

# Override standard setup! We don't yet trust anything
function setup() {
    :
}

@test "skopeo version emits reasonable output" {
    run_skopeo --version

    expect_output --substring "skopeo version [0-9.]+"
}

# vim: filetype=sh
