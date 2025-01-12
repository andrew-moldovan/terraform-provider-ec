// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package trafficfilterresource

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/elastic/cloud-sdk-go/pkg/models"
	"github.com/elastic/cloud-sdk-go/pkg/util/ec"
)

func Test_modelToState(t *testing.T) {
	remoteState := models.TrafficFilterRulesetInfo{
		ID:               ec.String("some-random-id"),
		Name:             ec.String("my traffic filter"),
		Type:             ec.String("ip"),
		IncludeByDefault: ec.Bool(false),
		Region:           ec.String("us-east-1"),
		Rules: []*models.TrafficFilterRule{
			{Source: "1.1.1.1"},
			{Source: "0.0.0.0/0"},
		},
	}

	remoteStateMultipleRules := models.TrafficFilterRulesetInfo{
		ID:               ec.String("some-random-id"),
		Name:             ec.String("my traffic filter"),
		Type:             ec.String("ip"),
		IncludeByDefault: ec.Bool(false),
		Region:           ec.String("us-east-1"),
		Rules: []*models.TrafficFilterRule{
			{Source: "1.1.1.0/16"},
			{Source: "1.1.1.1/24"},
			{Source: "0.0.0.0/0"},
			{Source: "1.1.1.1"},
		},
	}

	remoteStateMultipleRulesWithDesc := models.TrafficFilterRulesetInfo{
		ID:               ec.String("some-random-id"),
		Name:             ec.String("my traffic filter"),
		Type:             ec.String("ip"),
		IncludeByDefault: ec.Bool(false),
		Region:           ec.String("us-east-1"),
		Description:      *ec.String("Allows access to some network, a specific IP and all internet traffic"),
		Rules: []*models.TrafficFilterRule{
			{Source: "1.1.1.0/16", Description: "some network"},
			{Source: "1.1.1.1/24", Description: "a specific IP"},
			{Source: "0.0.0.0/0", Description: "all internet traffic"},
		},
	}

	want := newSampleTrafficFilter("some-random-id")
	wantMultipleRules := modelV0{
		ID:               types.String{Value: "some-random-id"},
		Name:             types.String{Value: "my traffic filter"},
		Type:             types.String{Value: "ip"},
		IncludeByDefault: types.Bool{Value: false},
		Region:           types.String{Value: "us-east-1"},
		Description:      types.String{Null: true},
		Rule: types.Set{
			ElemType: trafficFilterRuleElemType(),
			Elems: []attr.Value{
				newSampleTrafficFilterRule("1.1.1.0/16", "", "", "", ""),
				newSampleTrafficFilterRule("1.1.1.1/24", "", "", "", ""),
				newSampleTrafficFilterRule("0.0.0.0/0", "", "", "", ""),
				newSampleTrafficFilterRule("1.1.1.1", "", "", "", ""),
			},
		},
	}
	wantMultipleRulesWithDesc := modelV0{
		ID:               types.String{Value: "some-random-id"},
		Name:             types.String{Value: "my traffic filter"},
		Type:             types.String{Value: "ip"},
		IncludeByDefault: types.Bool{Value: false},
		Region:           types.String{Value: "us-east-1"},
		Description:      types.String{Value: "Allows access to some network, a specific IP and all internet traffic"},
		Rule: types.Set{
			ElemType: trafficFilterRuleElemType(),
			Elems: []attr.Value{
				newSampleTrafficFilterRule("1.1.1.0/16", "some network", "", "", ""),
				newSampleTrafficFilterRule("1.1.1.1/24", "a specific IP", "", "", ""),
				newSampleTrafficFilterRule("0.0.0.0/0", "all internet traffic", "", "", ""),
			},
		},
	}

	remoteStateAzurePL := models.TrafficFilterRulesetInfo{
		ID:               ec.String("some-random-id"),
		Name:             ec.String("my traffic filter"),
		Type:             ec.String("azure_private_endpoint"),
		IncludeByDefault: ec.Bool(false),
		Region:           ec.String("azure-australiaeast"),
		Rules: []*models.TrafficFilterRule{
			{
				AzureEndpointGUID: "1231312-1231-1231-1231-1231312",
				AzureEndpointName: "my-azure-pl",
			},
		},
	}

	wantAzurePL := modelV0{
		ID:               types.String{Value: "some-random-id"},
		Name:             types.String{Value: "my traffic filter"},
		Type:             types.String{Value: "azure_private_endpoint"},
		IncludeByDefault: types.Bool{Value: false},
		Region:           types.String{Value: "azure-australiaeast"},
		Description:      types.String{Null: true},
		Rule: types.Set{
			ElemType: trafficFilterRuleElemType(),
			Elems: []attr.Value{
				newSampleTrafficFilterRule("", "", "my-azure-pl", "1231312-1231-1231-1231-1231312", ""),
			},
		},
	}

	type args struct {
		in *models.TrafficFilterRulesetInfo
	}

	tests := []struct {
		name string
		args args
		err  error
		want modelV0
	}{
		{
			name: "flattens the resource",
			args: args{in: &remoteState},
			want: want,
		},
		{
			name: "flattens the resource with multiple rules",
			args: args{in: &remoteStateMultipleRules},
			want: wantMultipleRules,
		},
		{
			name: "flattens the resource with multiple rules with descriptions",
			args: args{in: &remoteStateMultipleRulesWithDesc},
			want: wantMultipleRulesWithDesc,
		},
		{
			name: "flattens the resource with multiple rules with descriptions",
			args: args{in: &remoteStateAzurePL},
			want: wantAzurePL,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := modelV0{
				ID: types.String{Value: "some-random-id"},
			}
			diags := modelToState(context.Background(), tt.args.in, &state)

			if tt.err != nil {
				assert.Equal(t, diags, tt.err)
			} else {
				assert.Empty(t, diags)
			}

			assert.Equal(t, tt.want, state)
		})
	}
}
