//go:build windows
// +build windows

/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package preflight

import (
	"os/user"

	"github.com/easysoft/qcadmin/internal/pkg/util/log"
	"github.com/pkg/errors"
)

// The "Well-known SID" of Administrator group
// https://support.microsoft.com/en-us/help/243330/well-known-security-identifiers-in-windows-operating-systems
const administratorSID = "S-1-5-32-544"

// Check validates if a user has elevated (administrator) privileges.
func (ipuc IsPrivilegedUserCheck) Check() error {
	currUser, err := user.Current()
	if err != nil {
		return errors.Wrap(err, "cannot get current user")
	}

	groupIds, err := currUser.GroupIds()
	if err != nil {
		return errors.Wrap(err, "cannot get group IDs for current user")
	}

	for _, sid := range groupIds {
		if sid == administratorSID {
			return nil
		}
	}

	return errors.New("user is not running as administrator")
}

// Check number of memory required by kubeadm
// No-op for Windows.
func (mc MemCheck) Check() error {
	log.Flog.Warnf("validating number of Memory is not supported on Windows, Skipping")
	return nil
}
