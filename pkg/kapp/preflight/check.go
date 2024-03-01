// Copyright 2024 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package preflight

import (
	"context"

	ctldgraph "github.com/vmware-tanzu/carvel-kapp/pkg/kapp/diffgraph"
)

type CheckFunc func(context.Context, *ctldgraph.ChangeGraph) error

type Check interface {
	Enabled() bool
	SetEnabled(bool)
	SetConfig(map[string]any) error
	Run(context.Context, *ctldgraph.ChangeGraph) error
}

type checkImpl struct {
	enabled   bool
	checkFunc CheckFunc
	config    map[string]any
}

func NewCheck(cf CheckFunc, enabled bool) Check {
	return &checkImpl{
		enabled:   enabled,
		checkFunc: cf,
	}
}

func (cf *checkImpl) Enabled() bool {
	return cf.enabled
}

func (cf *checkImpl) SetEnabled(enabled bool) {
	cf.enabled = enabled
}

func (cf *checkImpl) Run(ctx context.Context, changeGraph *ctldgraph.ChangeGraph) error {
	return cf.checkFunc(ctx, changeGraph)
}

func (cf *checkImpl) SetConfig(config map[string]any) error {
	cf.config = config
	return nil
}
