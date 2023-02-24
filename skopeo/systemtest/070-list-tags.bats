#!/usr/bin/env bats
#
# list-tags tests
#

load helpers

# list from registry
@test "list-tags: remote repository on a registry" {
    local remote_image=quay.io/libpod/alpine_labels

    run_skopeo list-tags "docker://${remote_image}"
    expect_output --substring "quay.io/libpod/alpine_labels"
    expect_output --substring "latest"
}

# list from a local docker-archive file
@test "list-tags: from a docker-archive file" {
    local file_name=${TEST_SOURCE_DIR}/testdata/docker-two-images.tar.xz

    run_skopeo list-tags docker-archive:$file_name
    expect_output --substring "example.com/empty:latest"
    expect_output --substring "example.com/empty/but:different"

}


# vim: filetype=sh
