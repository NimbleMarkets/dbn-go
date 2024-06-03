// Copyright (c) 2024 Neomantra Corp

package dbn

// NullVisitor is an implementation of all the dbn.Visitor interface.
// It is useful for copy/pasting to ones own implementation.
type NullVisitor struct {
}

func (v *NullVisitor) OnMbp0(record *Mbp0Msg) error {
	return nil
}

func (v *NullVisitor) OnMbp10(record *Mbp10Msg) error {
	return nil
}

func (v *NullVisitor) OnMbp1(record *Mbp1Msg) error {
	return nil
}

func (v *NullVisitor) OnMbo(record *MboMsg) error {
	return nil
}

func (v *NullVisitor) OnOhlcv(record *OhlcvMsg) error {
	return nil
}

func (v *NullVisitor) OnImbalance(record *ImbalanceMsg) error {
	return nil
}

func (v *NullVisitor) OnStatMsg(record *StatMsg) error {
	return nil
}

func (v *NullVisitor) OnErrorMsg(record *ErrorMsg) error {
	return nil
}

func (v *NullVisitor) OnSystemMsg(record *SystemMsg) error {
	return nil
}

func (v *NullVisitor) OnSymbolMappingMsg(record *SymbolMappingMsg) error {
	return nil
}

func (v *NullVisitor) OnStreamEnd() error {
	return nil
}
