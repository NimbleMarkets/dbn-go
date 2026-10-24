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
		&dbn.InstrumentDefMsgV2{}, &dbn.InstrumentDefMsgV3{},
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
					v, err := fastjson.Parse(`{"levels":[{},{},{},{},{},{},{},{},{},{}],"action":` + tc.input + `,"side":` + tc.input + `,"auction_type":` + tc.input + `,"unpaired_side":` + tc.input + `,"significant_imbalance":` + tc.input + `,"is_trading":` + tc.input + `,"is_quoting":` + tc.input + `,"is_short_sell_restricted":` + tc.input + `,"instrument_class":` + tc.input + `,"match_algorithm":` + tc.input + `,"security_update_action":` + tc.input + `,"user_defined_instrument":` + tc.input + `,"leg_instrument_class":` + tc.input + `,"leg_side":` + tc.input + `}`)
					if err != nil {
						t.Fatal(err)
					}
					if err := record.Fill_Json(v, &dbn.RHeader{}); err != nil {
						t.Fatal(err)
					}
					fields := []string{"Action", "Side"}
					if _, ok := record.(*dbn.ImbalanceMsg); ok {
						fields = []string{"Side", "AuctionType", "UnpairedSide", "SignificantImbalance"}
					}
					if _, ok := record.(*dbn.StatusMsg); ok {
						fields = []string{"IsTrading", "IsQuoting", "IsShortSellRestricted"}
					}
					if _, ok := record.(*dbn.InstrumentDefMsgV2); ok {
						fields = []string{"InstrumentClass", "MatchAlgorithm", "SecurityUpdateAction", "UserDefinedInstrument"}
					}
					if _, ok := record.(*dbn.InstrumentDefMsgV3); ok {
						fields = []string{"InstrumentClass", "MatchAlgorithm", "SecurityUpdateAction", "UserDefinedInstrument", "LegInstrumentClass", "LegSide"}
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

func TestImbalanceJSONCodes(t *testing.T) {
	for _, input := range []string{
		`{"side":"A","auction_type":"O","unpaired_side":"N","significant_imbalance":"~"}`,
		`{"side":65,"auction_type":79,"unpaired_side":78,"significant_imbalance":126}`,
	} {
		v, err := fastjson.Parse(input)
		if err != nil {
			t.Fatal(err)
		}
		var record dbn.ImbalanceMsg
		if err := record.Fill_Json(v, &dbn.RHeader{}); err != nil {
			t.Fatal(err)
		}
		got := [4]byte{record.Side, record.AuctionType, record.UnpairedSide, record.SignificantImbalance}
		if want := [4]byte{'A', 'O', 'N', '~'}; got != want {
			t.Errorf("decoded codes = %q, want %q", got, want)
		}
	}
}
