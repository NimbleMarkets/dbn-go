// Copyright (c) 2024-2025 Neomantra Corp

package dbn_test

import (
	"reflect"
	"testing"

	dbn "github.com/NimbleMarkets/dbn-go"
	"github.com/valyala/fastjson"
)

func TestJSONCharacterEncodings(t *testing.T) {
	for _, record := range []interface {
		Fill_Json(*fastjson.Value, *dbn.RHeader) error
	}{
		&dbn.MboMsg{}, &dbn.Mbp0Msg{}, &dbn.Mbp1Msg{}, &dbn.Mbp10Msg{},
		&dbn.Cmbp1Msg{}, &dbn.BboMsg{}, &dbn.ImbalanceMsg{}, &dbn.StatusMsg{},
	} {
		t.Run(reflect.TypeOf(record).Elem().Name(), func(t *testing.T) {
			for _, tc := range []struct {
				name, input string
				want        byte
			}{
				{"string", `"A"`, 'A'},
				{"numeric", `65`, 'A'},
				{"escaped", `"\u0041"`, 'A'},
				{"unset", `"~"`, '~'},
				{"empty", `""`, 0},
				{"null", `null`, 0},
				{"multi-character", `"AB"`, 0},
			} {
				t.Run(tc.name, func(t *testing.T) {
					v, err := fastjson.Parse(`{"levels":[{},{},{},{},{},{},{},{},{},{}],"action":` + tc.input + `,"side":` + tc.input + `,"is_trading":` + tc.input + `,"is_quoting":` + tc.input + `,"is_short_sell_restricted":` + tc.input + `}`)
					if err != nil {
						t.Fatal(err)
					}
					if err := record.Fill_Json(v, &dbn.RHeader{}); err != nil {
						t.Fatal(err)
					}
					fields := []string{"Action", "Side"}
					if _, ok := record.(*dbn.StatusMsg); ok {
						fields = []string{"IsTrading", "IsQuoting", "IsShortSellRestricted"}
					}
					for _, name := range fields {
						field := reflect.ValueOf(record).Elem().FieldByName(name)
						if field.IsValid() && field.Uint() != uint64(tc.want) {
							t.Errorf("%s = %d, want %d", name, field.Uint(), tc.want)
						}
					}
				})
			}
		})
	}
}
