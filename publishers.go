// Copyright (c) 2024-2025 Neomantra Corp
//
// Adapted from Databento's DBN:
//   https://github.com/databento/dbn/blob/main/rust/dbn/src/publishers.rs
//

package dbn

import (
	"encoding/json"
	"fmt"
	"strings"
)

///////////////////////////////////////////////////////////////////////////////
// Venue
///////////////////////////////////////////////////////////////////////////////

// Venue is a trading execution venue.
type Venue uint16

const (
	// CME Globex
	Venue_Glbx Venue = 1
	// Nasdaq - All Markets
	Venue_Xnas Venue = 2
	// Nasdaq OMX BX
	Venue_Xbos Venue = 3
	// Nasdaq OMX PSX
	Venue_Xpsx Venue = 4
	// Cboe BZX U.S. Equities Exchange
	Venue_Bats Venue = 5
	// Cboe BYX U.S. Equities Exchange
	Venue_Baty Venue = 6
	// Cboe EDGA U.S. Equities Exchange
	Venue_Edga Venue = 7
	// Cboe EDGX U.S. Equities Exchange
	Venue_Edgx Venue = 8
	// New York Stock Exchange, Inc.
	Venue_Xnys Venue = 9
	// NYSE National, Inc.
	Venue_Xcis Venue = 10
	// NYSE MKT LLC
	Venue_Xase Venue = 11
	// NYSE Arca
	Venue_Arcx Venue = 12
	// NYSE Texas, Inc.
	Venue_Xchi Venue = 13
	// Investors Exchange
	Venue_Iexg Venue = 14
	// FINRA/Nasdaq TRF Carteret
	Venue_Finn Venue = 15
	// FINRA/Nasdaq TRF Chicago
	Venue_Finc Venue = 16
	// FINRA/NYSE TRF
	Venue_Finy Venue = 17
	// MEMX LLC Equities
	Venue_Memx Venue = 18
	// MIAX Pearl Equities
	Venue_Eprl Venue = 19
	// NYSE American Options
	Venue_Amxo Venue = 20
	// BOX Options
	Venue_Xbox Venue = 21
	// Cboe Options
	Venue_Xcbo Venue = 22
	// MIAX Emerald
	Venue_Emld Venue = 23
	// Cboe EDGX Options
	Venue_Edgo Venue = 24
	// Nasdaq GEMX
	Venue_Gmni Venue = 25
	// Nasdaq ISE
	Venue_Xisx Venue = 26
	// Nasdaq MRX
	Venue_Mcry Venue = 27
	// MIAX Options
	Venue_Xmio Venue = 28
	// NYSE Arca Options
	Venue_Arco Venue = 29
	// Options Price Reporting Authority
	Venue_Opra Venue = 30
	// MIAX Pearl
	Venue_Mprl Venue = 31
	// Nasdaq Options
	Venue_Xndq Venue = 32
	// Nasdaq Texas Options
	Venue_Xbxo Venue = 33
	// Cboe C2 Options
	Venue_C2Ox Venue = 34
	// Nasdaq PHLX
	Venue_Xphl Venue = 35
	// Cboe BZX Options
	Venue_Bato Venue = 36
	// MEMX Options
	Venue_Mxop Venue = 37
	// ICE Europe Commodities
	Venue_Ifeu Venue = 38
	// ICE Endex
	Venue_Ndex Venue = 39
	// Databento US Equities - Consolidated
	Venue_Dbeq Venue = 40
	// MIAX Sapphire
	Venue_Sphr Venue = 41
	// Long-Term Stock Exchange, Inc.
	Venue_Ltse Venue = 42
	// Off-Exchange Transactions - Listed Instruments
	Venue_Xoff Venue = 43
	// IntelligentCross ASPEN Intelligent Bid/Offer
	Venue_Aspn Venue = 44
	// IntelligentCross ASPEN Maker/Taker
	Venue_Asmt Venue = 45
	// IntelligentCross ASPEN Inverted
	Venue_Aspi Venue = 46
	// Databento US Equities - Consolidated
	Venue_Equs Venue = 47
	// ICE Futures US
	Venue_Ifus Venue = 48
	// ICE Europe Financials
	Venue_Ifll Venue = 49
	// Eurex Exchange
	Venue_Xeur Venue = 50
	// European Energy Exchange
	Venue_Xeee Venue = 51
	// Cboe Futures Exchange
	Venue_Xcbf Venue = 52
	// Blue Ocean ATS
	Venue_Ocea Venue = 53
	// MX2 Options
	Venue_Mxto Venue = 54
	// IEX Options LLC
	Venue_Iexo Venue = 55
	// Cboe Global Indices Feed
	Venue_Main Venue = 56
	// Texas Stock Exchange, LLC
	Venue_Txse Venue = 57
	// 24X National Exchange, LLC
	Venue__24Eq Venue = 58
	// Consolidated Tape Association
	Venue_Cta Venue = 59
	// Unlisted Trading Privileges
	Venue_Utp Venue = 60
	// FINRA Alternative Display Facility
	Venue_Xadf Venue = 61
	// CGIF Morgan Stanley Capital International
	Venue_Msci Venue = 62
	// CGIF FTSE Russell
	Venue_Ftse Venue = 63
	// CGIF Intraday Net Asset Values
	Venue_Inav Venue = 64
	// CGIF Morning Star
	Venue_Mstar Venue = 65
	// CGIF Crypto Currency
	Venue_Cccy Venue = 66
	// CGIF CBOE Global Indices
	Venue_Cgi Venue = 67
	// CGIF Definitions
	Venue_Cdef Venue = 68
	// Tokyo Stock Exchange
	Venue_Xtks Venue = 69
	// Tokyo Commodity Exchange
	Venue_Xtkt Venue = 70
	// Osaka Exchange
	Venue_Xose Venue = 71
	// Databento Core Indices
	Venue_Dbix Venue = 72
)

const VENUE_COUNT = 72

// Returns the string representation of the Venue, or empty string if unknown.
func (v Venue) String() string {
	switch v {
	case Venue_Glbx:
		return "GLBX"
	case Venue_Xnas:
		return "XNAS"
	case Venue_Xbos:
		return "XBOS"
	case Venue_Xpsx:
		return "XPSX"
	case Venue_Bats:
		return "BATS"
	case Venue_Baty:
		return "BATY"
	case Venue_Edga:
		return "EDGA"
	case Venue_Edgx:
		return "EDGX"
	case Venue_Xnys:
		return "XNYS"
	case Venue_Xcis:
		return "XCIS"
	case Venue_Xase:
		return "XASE"
	case Venue_Arcx:
		return "ARCX"
	case Venue_Xchi:
		return "XCHI"
	case Venue_Iexg:
		return "IEXG"
	case Venue_Finn:
		return "FINN"
	case Venue_Finc:
		return "FINC"
	case Venue_Finy:
		return "FINY"
	case Venue_Memx:
		return "MEMX"
	case Venue_Eprl:
		return "EPRL"
	case Venue_Amxo:
		return "AMXO"
	case Venue_Xbox:
		return "XBOX"
	case Venue_Xcbo:
		return "XCBO"
	case Venue_Emld:
		return "EMLD"
	case Venue_Edgo:
		return "EDGO"
	case Venue_Gmni:
		return "GMNI"
	case Venue_Xisx:
		return "XISX"
	case Venue_Mcry:
		return "MCRY"
	case Venue_Xmio:
		return "XMIO"
	case Venue_Arco:
		return "ARCO"
	case Venue_Opra:
		return "OPRA"
	case Venue_Mprl:
		return "MPRL"
	case Venue_Xndq:
		return "XNDQ"
	case Venue_Xbxo:
		return "XBXO"
	case Venue_C2Ox:
		return "C2OX"
	case Venue_Xphl:
		return "XPHL"
	case Venue_Bato:
		return "BATO"
	case Venue_Mxop:
		return "MXOP"
	case Venue_Ifeu:
		return "IFEU"
	case Venue_Ndex:
		return "NDEX"
	case Venue_Dbeq:
		return "DBEQ"
	case Venue_Sphr:
		return "SPHR"
	case Venue_Ltse:
		return "LTSE"
	case Venue_Xoff:
		return "XOFF"
	case Venue_Aspn:
		return "ASPN"
	case Venue_Asmt:
		return "ASMT"
	case Venue_Aspi:
		return "ASPI"
	case Venue_Equs:
		return "EQUS"
	case Venue_Ifus:
		return "IFUS"
	case Venue_Ifll:
		return "IFLL"
	case Venue_Xeur:
		return "XEUR"
	case Venue_Xeee:
		return "XEEE"
	case Venue_Xcbf:
		return "XCBF"
	case Venue_Ocea:
		return "OCEA"
	case Venue_Mxto:
		return "MXTO"
	case Venue_Iexo:
		return "IEXO"
	case Venue_Main:
		return "MAIN"
	case Venue_Txse:
		return "TXSE"
	case Venue__24Eq:
		return "24EQ"
	case Venue_Cta:
		return "CTA"
	case Venue_Utp:
		return "UTP"
	case Venue_Xadf:
		return "XADF"
	case Venue_Msci:
		return "MSCI"
	case Venue_Ftse:
		return "FTSE"
	case Venue_Inav:
		return "INAV"
	case Venue_Mstar:
		return "MSTAR"
	case Venue_Cccy:
		return "CCCY"
	case Venue_Cgi:
		return "CGI"
	case Venue_Cdef:
		return "CDEF"
	case Venue_Xtks:
		return "XTKS"
	case Venue_Xtkt:
		return "XTKT"
	case Venue_Xose:
		return "XOSE"
	case Venue_Dbix:
		return "DBIX"
	default:
		return ""
	}
}

// VenueFromString converts a string to a Venue.
// Returns an error if the string is unknown.
func VenueFromString(str string) (Venue, error) {
	str = strings.ToUpper(str)
	switch str {
	case "GLBX":
		return Venue_Glbx, nil
	case "XNAS":
		return Venue_Xnas, nil
	case "XBOS":
		return Venue_Xbos, nil
	case "XPSX":
		return Venue_Xpsx, nil
	case "BATS":
		return Venue_Bats, nil
	case "BATY":
		return Venue_Baty, nil
	case "EDGA":
		return Venue_Edga, nil
	case "EDGX":
		return Venue_Edgx, nil
	case "XNYS":
		return Venue_Xnys, nil
	case "XCIS":
		return Venue_Xcis, nil
	case "XASE":
		return Venue_Xase, nil
	case "ARCX":
		return Venue_Arcx, nil
	case "XCHI":
		return Venue_Xchi, nil
	case "IEXG":
		return Venue_Iexg, nil
	case "FINN":
		return Venue_Finn, nil
	case "FINC":
		return Venue_Finc, nil
	case "FINY":
		return Venue_Finy, nil
	case "MEMX":
		return Venue_Memx, nil
	case "EPRL":
		return Venue_Eprl, nil
	case "AMXO":
		return Venue_Amxo, nil
	case "XBOX":
		return Venue_Xbox, nil
	case "XCBO":
		return Venue_Xcbo, nil
	case "EMLD":
		return Venue_Emld, nil
	case "EDGO":
		return Venue_Edgo, nil
	case "GMNI":
		return Venue_Gmni, nil
	case "XISX":
		return Venue_Xisx, nil
	case "MCRY":
		return Venue_Mcry, nil
	case "XMIO":
		return Venue_Xmio, nil
	case "ARCO":
		return Venue_Arco, nil
	case "OPRA":
		return Venue_Opra, nil
	case "MPRL":
		return Venue_Mprl, nil
	case "XNDQ":
		return Venue_Xndq, nil
	case "XBXO":
		return Venue_Xbxo, nil
	case "C2OX":
		return Venue_C2Ox, nil
	case "XPHL":
		return Venue_Xphl, nil
	case "BATO":
		return Venue_Bato, nil
	case "MXOP":
		return Venue_Mxop, nil
	case "IFEU":
		return Venue_Ifeu, nil
	case "NDEX":
		return Venue_Ndex, nil
	case "DBEQ":
		return Venue_Dbeq, nil
	case "SPHR":
		return Venue_Sphr, nil
	case "LTSE":
		return Venue_Ltse, nil
	case "XOFF":
		return Venue_Xoff, nil
	case "ASPN":
		return Venue_Aspn, nil
	case "ASMT":
		return Venue_Asmt, nil
	case "ASPI":
		return Venue_Aspi, nil
	case "EQUS":
		return Venue_Equs, nil
	case "IFUS":
		return Venue_Ifus, nil
	case "IFLL":
		return Venue_Ifll, nil
	case "XEUR":
		return Venue_Xeur, nil
	case "XEEE":
		return Venue_Xeee, nil
	case "XCBF":
		return Venue_Xcbf, nil
	case "OCEA":
		return Venue_Ocea, nil
	case "MXTO":
		return Venue_Mxto, nil
	case "IEXO":
		return Venue_Iexo, nil
	case "MAIN":
		return Venue_Main, nil
	case "TXSE":
		return Venue_Txse, nil
	case "24EQ":
		return Venue__24Eq, nil
	case "CTA":
		return Venue_Cta, nil
	case "UTP":
		return Venue_Utp, nil
	case "XADF":
		return Venue_Xadf, nil
	case "MSCI":
		return Venue_Msci, nil
	case "FTSE":
		return Venue_Ftse, nil
	case "INAV":
		return Venue_Inav, nil
	case "MSTAR":
		return Venue_Mstar, nil
	case "CCCY":
		return Venue_Cccy, nil
	case "CGI":
		return Venue_Cgi, nil
	case "CDEF":
		return Venue_Cdef, nil
	case "XTKS":
		return Venue_Xtks, nil
	case "XTKT":
		return Venue_Xtkt, nil
	case "XOSE":
		return Venue_Xose, nil
	case "DBIX":
		return Venue_Dbix, nil
	default:
		return Venue_Glbx, fmt.Errorf("unknown venue: '%s'", str)
	}
}

func (v Venue) MarshalJSON() ([]byte, error) {
	return json.Marshal(v.String())
}

func (v *Venue) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	vn, err := VenueFromString(str)
	if err != nil {
		return err
	}
	*v = vn
	return nil
}

// Type implements pflag.Value.Type. Returns "dbn.Venue".
func (*Venue) Type() string {
	return "dbn.Venue"
}

// Set implements the flag.Value interface.
func (v *Venue) Set(value string) error {
	vn, err := VenueFromString(value)
	if err == nil {
		*v = vn
	}
	return err
}

///////////////////////////////////////////////////////////////////////////////
// Dataset
///////////////////////////////////////////////////////////////////////////////

// Dataset is a source of data.
type Dataset uint16

const (
	// CME MDP 3.0 Market Data
	Dataset_GlbxMdp3 Dataset = 1
	// Nasdaq TotalView-ITCH
	Dataset_XnasItch Dataset = 2
	// Nasdaq Texas TotalView-ITCH
	Dataset_XbosItch Dataset = 3
	// Nasdaq PSX TotalView-ITCH
	Dataset_XpsxItch Dataset = 4
	// Cboe BZX Depth
	Dataset_BatsPitch Dataset = 5
	// Cboe BYX Depth
	Dataset_BatyPitch Dataset = 6
	// Cboe EDGA Depth
	Dataset_EdgaPitch Dataset = 7
	// Cboe EDGX Depth
	Dataset_EdgxPitch Dataset = 8
	// NYSE Integrated
	Dataset_XnysPillar Dataset = 9
	// NYSE National Integrated
	Dataset_XcisPillar Dataset = 10
	// NYSE American Integrated
	Dataset_XasePillar Dataset = 11
	// NYSE Texas Integrated
	Dataset_XchiPillar Dataset = 12
	// NYSE National BBO
	Dataset_XcisBbo Dataset = 13
	// NYSE National Trades
	Dataset_XcisTrades Dataset = 14
	// MEMX MEMOIR Depth
	Dataset_MemxMemoir Dataset = 15
	// MIAX Pearl Depth
	Dataset_EprlDom Dataset = 16
	// FINRA/Nasdaq TRF (DEPRECATED)
	Dataset_FinnNls Dataset = 17
	// FINRA/NYSE TRF (DEPRECATED)
	Dataset_FinyTrades Dataset = 18
	// OPRA Binary
	Dataset_OpraPillar Dataset = 19
	// Databento US Equities Basic
	Dataset_DbeqBasic Dataset = 20
	// NYSE Arca Integrated
	Dataset_ArcxPillar Dataset = 21
	// IEX TOPS
	Dataset_IexgTops Dataset = 22
	// Databento US Equities Plus
	Dataset_EqusPlus Dataset = 23
	// NYSE BBO
	Dataset_XnysBbo Dataset = 24
	// NYSE Trades
	Dataset_XnysTrades Dataset = 25
	// Nasdaq QBBO
	Dataset_XnasQbbo Dataset = 26
	// Nasdaq NLS
	Dataset_XnasNls Dataset = 27
	// ICE Europe Commodities iMpact
	Dataset_IfeuImpact Dataset = 28
	// ICE Endex iMpact
	Dataset_NdexImpact Dataset = 29
	// Databento US Equities (All Feeds)
	Dataset_EqusAll Dataset = 30
	// Nasdaq Basic (NLS and QBBO)
	Dataset_XnasBasic Dataset = 31
	// Databento US Equities Summary
	Dataset_EqusSummary Dataset = 32
	// NYSE National Trades and BBO
	Dataset_XcisTradesbbo Dataset = 33
	// NYSE Trades and BBO
	Dataset_XnysTradesbbo Dataset = 34
	// Databento US Equities Mini
	Dataset_EqusMini Dataset = 35
	// ICE Futures US iMpact
	Dataset_IfusImpact Dataset = 36
	// ICE Europe Financials iMpact
	Dataset_IfllImpact Dataset = 37
	// Eurex EOBI
	Dataset_XeurEobi Dataset = 38
	// European Energy Exchange EOBI
	Dataset_XeeeEobi Dataset = 39
	// CFE Depth
	Dataset_XcbfPitch Dataset = 40
	// Blue Ocean ATS MEMOIR Depth
	Dataset_OceaMemoir Dataset = 41
	// Cboe Titanium Cboe Global Indices Feed
	Dataset_MainCgif Dataset = 42
	// US Equities Security Information Processor
	Dataset_EqusSip Dataset = 43
	// CGIF - Morgan Stanley Capital International
	Dataset_MsciCgif Dataset = 44
	// CGIF - FTSE Russell
	Dataset_FtseCgif Dataset = 45
	// CGIF - Intraday Net Asset Values
	Dataset_InavCgif Dataset = 46
	// CGIF - Morning Star
	Dataset_MstarCgif Dataset = 47
	// CGIF - Crypto Currency
	Dataset_CccyCgif Dataset = 48
	// CGIF - CBOE Global Indices
	Dataset_CgiCgif Dataset = 49
	// Arrowhead FLEX
	Dataset_XtksFlex Dataset = 50
	// JPX JGATE3.0-ITCH
	Dataset_XtktItch Dataset = 51
	// JPX JGATE3.0-ITCH
	Dataset_XoseItch Dataset = 52
	// Databento Core Indices
	Dataset_DbixCore Dataset = 53
)

const DATASET_COUNT = 53

// Returns the string representation of the Dataset, or empty string if unknown.
func (d Dataset) String() string {
	switch d {
	case Dataset_GlbxMdp3:
		return "GLBX.MDP3"
	case Dataset_XnasItch:
		return "XNAS.ITCH"
	case Dataset_XbosItch:
		return "XBOS.ITCH"
	case Dataset_XpsxItch:
		return "XPSX.ITCH"
	case Dataset_BatsPitch:
		return "BATS.PITCH"
	case Dataset_BatyPitch:
		return "BATY.PITCH"
	case Dataset_EdgaPitch:
		return "EDGA.PITCH"
	case Dataset_EdgxPitch:
		return "EDGX.PITCH"
	case Dataset_XnysPillar:
		return "XNYS.PILLAR"
	case Dataset_XcisPillar:
		return "XCIS.PILLAR"
	case Dataset_XasePillar:
		return "XASE.PILLAR"
	case Dataset_XchiPillar:
		return "XCHI.PILLAR"
	case Dataset_XcisBbo:
		return "XCIS.BBO"
	case Dataset_XcisTrades:
		return "XCIS.TRADES"
	case Dataset_MemxMemoir:
		return "MEMX.MEMOIR"
	case Dataset_EprlDom:
		return "EPRL.DOM"
	case Dataset_FinnNls:
		return "FINN.NLS"
	case Dataset_FinyTrades:
		return "FINY.TRADES"
	case Dataset_OpraPillar:
		return "OPRA.PILLAR"
	case Dataset_DbeqBasic:
		return "DBEQ.BASIC"
	case Dataset_ArcxPillar:
		return "ARCX.PILLAR"
	case Dataset_IexgTops:
		return "IEXG.TOPS"
	case Dataset_EqusPlus:
		return "EQUS.PLUS"
	case Dataset_XnysBbo:
		return "XNYS.BBO"
	case Dataset_XnysTrades:
		return "XNYS.TRADES"
	case Dataset_XnasQbbo:
		return "XNAS.QBBO"
	case Dataset_XnasNls:
		return "XNAS.NLS"
	case Dataset_IfeuImpact:
		return "IFEU.IMPACT"
	case Dataset_NdexImpact:
		return "NDEX.IMPACT"
	case Dataset_EqusAll:
		return "EQUS.ALL"
	case Dataset_XnasBasic:
		return "XNAS.BASIC"
	case Dataset_EqusSummary:
		return "EQUS.SUMMARY"
	case Dataset_XcisTradesbbo:
		return "XCIS.TRADESBBO"
	case Dataset_XnysTradesbbo:
		return "XNYS.TRADESBBO"
	case Dataset_EqusMini:
		return "EQUS.MINI"
	case Dataset_IfusImpact:
		return "IFUS.IMPACT"
	case Dataset_IfllImpact:
		return "IFLL.IMPACT"
	case Dataset_XeurEobi:
		return "XEUR.EOBI"
	case Dataset_XeeeEobi:
		return "XEEE.EOBI"
	case Dataset_XcbfPitch:
		return "XCBF.PITCH"
	case Dataset_OceaMemoir:
		return "OCEA.MEMOIR"
	case Dataset_MainCgif:
		return "MAIN.CGIF"
	case Dataset_EqusSip:
		return "EQUS.SIP"
	case Dataset_MsciCgif:
		return "MSCI.CGIF"
	case Dataset_FtseCgif:
		return "FTSE.CGIF"
	case Dataset_InavCgif:
		return "INAV.CGIF"
	case Dataset_MstarCgif:
		return "MSTAR.CGIF"
	case Dataset_CccyCgif:
		return "CCCY.CGIF"
	case Dataset_CgiCgif:
		return "CGI.CGIF"
	case Dataset_XtksFlex:
		return "XTKS.FLEX"
	case Dataset_XtktItch:
		return "XTKT.ITCH"
	case Dataset_XoseItch:
		return "XOSE.ITCH"
	case Dataset_DbixCore:
		return "DBIX.CORE"
	default:
		return ""
	}
}

// Publishers returns all Publisher values associated with this dataset.
// Tracks https://github.com/databento/dbn/blob/main/rust/dbn/src/publishers.rs#L410
func (d Dataset) Publishers() []Publisher {
	switch d {
	case Dataset_GlbxMdp3:
		return []Publisher{Publisher_GlbxMdp3Glbx}
	case Dataset_XnasItch:
		return []Publisher{Publisher_XnasItchXnas}
	case Dataset_XbosItch:
		return []Publisher{Publisher_XbosItchXbos}
	case Dataset_XpsxItch:
		return []Publisher{Publisher_XpsxItchXpsx}
	case Dataset_BatsPitch:
		return []Publisher{Publisher_BatsPitchBats}
	case Dataset_BatyPitch:
		return []Publisher{Publisher_BatyPitchBaty}
	case Dataset_EdgaPitch:
		return []Publisher{Publisher_EdgaPitchEdga}
	case Dataset_EdgxPitch:
		return []Publisher{Publisher_EdgxPitchEdgx}
	case Dataset_XnysPillar:
		return []Publisher{Publisher_XnysPillarXnys}
	case Dataset_XcisPillar:
		return []Publisher{Publisher_XcisPillarXcis}
	case Dataset_XasePillar:
		return []Publisher{Publisher_XasePillarXase}
	case Dataset_XchiPillar:
		return []Publisher{Publisher_XchiPillarXchi}
	case Dataset_XcisBbo:
		return []Publisher{Publisher_XcisBboXcis}
	case Dataset_XcisTrades:
		return []Publisher{Publisher_XcisTradesXcis}
	case Dataset_MemxMemoir:
		return []Publisher{Publisher_MemxMemoirMemx}
	case Dataset_EprlDom:
		return []Publisher{Publisher_EprlDomEprl}
	case Dataset_FinnNls:
		return []Publisher{}
	case Dataset_FinyTrades:
		return []Publisher{}
	case Dataset_OpraPillar:
		return []Publisher{Publisher_OpraPillarAmxo, Publisher_OpraPillarXbox, Publisher_OpraPillarXcbo, Publisher_OpraPillarEmld, Publisher_OpraPillarEdgo, Publisher_OpraPillarGmni, Publisher_OpraPillarXisx, Publisher_OpraPillarMcry, Publisher_OpraPillarXmio, Publisher_OpraPillarArco, Publisher_OpraPillarOpra, Publisher_OpraPillarMprl, Publisher_OpraPillarXndq, Publisher_OpraPillarXbxo, Publisher_OpraPillarC2Ox, Publisher_OpraPillarXphl, Publisher_OpraPillarBato, Publisher_OpraPillarMxop, Publisher_OpraPillarSphr, Publisher_OpraPillarMxto, Publisher_OpraPillarIexo}
	case Dataset_DbeqBasic:
		return []Publisher{Publisher_DbeqBasicXchi, Publisher_DbeqBasicXcis, Publisher_DbeqBasicIexg, Publisher_DbeqBasicEprl, Publisher_DbeqBasicDbeq}
	case Dataset_ArcxPillar:
		return []Publisher{Publisher_ArcxPillarArcx}
	case Dataset_IexgTops:
		return []Publisher{Publisher_IexgTopsIexg}
	case Dataset_EqusPlus:
		return []Publisher{Publisher_EqusPlusXchi, Publisher_EqusPlusXcis, Publisher_EqusPlusIexg, Publisher_EqusPlusEprl, Publisher_EqusPlusXnas, Publisher_EqusPlusXnys, Publisher_EqusPlusFinn, Publisher_EqusPlusFiny, Publisher_EqusPlusFinc, Publisher_EqusPlusEqus}
	case Dataset_XnysBbo:
		return []Publisher{Publisher_XnysBboXnys}
	case Dataset_XnysTrades:
		return []Publisher{Publisher_XnysTradesFiny, Publisher_XnysTradesXnys, Publisher_XnysTradesEqus}
	case Dataset_XnasQbbo:
		return []Publisher{Publisher_XnasQbboXnas}
	case Dataset_XnasNls:
		return []Publisher{Publisher_XnasNlsFinn, Publisher_XnasNlsFinc, Publisher_XnasNlsXnas, Publisher_XnasNlsXbos, Publisher_XnasNlsXpsx}
	case Dataset_IfeuImpact:
		return []Publisher{Publisher_IfeuImpactIfeu, Publisher_IfeuImpactXoff}
	case Dataset_NdexImpact:
		return []Publisher{Publisher_NdexImpactNdex, Publisher_NdexImpactXoff}
	case Dataset_EqusAll:
		return []Publisher{Publisher_EqusAllXchi, Publisher_EqusAllXcis, Publisher_EqusAllIexg, Publisher_EqusAllEprl, Publisher_EqusAllXnas, Publisher_EqusAllXnys, Publisher_EqusAllFinn, Publisher_EqusAllFiny, Publisher_EqusAllFinc, Publisher_EqusAllBats, Publisher_EqusAllBaty, Publisher_EqusAllEdga, Publisher_EqusAllEdgx, Publisher_EqusAllXbos, Publisher_EqusAllXpsx, Publisher_EqusAllMemx, Publisher_EqusAllXase, Publisher_EqusAllArcx, Publisher_EqusAllLtse, Publisher_EqusAllEqus}
	case Dataset_XnasBasic:
		return []Publisher{Publisher_XnasBasicXnas, Publisher_XnasBasicFinn, Publisher_XnasBasicFinc, Publisher_XnasBasicXbos, Publisher_XnasBasicXpsx, Publisher_XnasBasicEqus}
	case Dataset_EqusSummary:
		return []Publisher{Publisher_EqusSummaryEqus}
	case Dataset_XcisTradesbbo:
		return []Publisher{Publisher_XcisTradesbboXcis}
	case Dataset_XnysTradesbbo:
		return []Publisher{Publisher_XnysTradesbboXnys}
	case Dataset_EqusMini:
		return []Publisher{Publisher_EqusMiniEqus}
	case Dataset_IfusImpact:
		return []Publisher{Publisher_IfusImpactIfus, Publisher_IfusImpactXoff}
	case Dataset_IfllImpact:
		return []Publisher{Publisher_IfllImpactIfll, Publisher_IfllImpactXoff}
	case Dataset_XeurEobi:
		return []Publisher{Publisher_XeurEobiXeur, Publisher_XeurEobiXoff}
	case Dataset_XeeeEobi:
		return []Publisher{Publisher_XeeeEobiXeee, Publisher_XeeeEobiXoff}
	case Dataset_XcbfPitch:
		return []Publisher{Publisher_XcbfPitchXcbf, Publisher_XcbfPitchXoff}
	case Dataset_OceaMemoir:
		return []Publisher{Publisher_OceaMemoirOcea}
	case Dataset_MainCgif:
		return []Publisher{Publisher_MainCgifMain, Publisher_MainCgifCdef}
	case Dataset_EqusSip:
		return []Publisher{Publisher_EqusSipXase, Publisher_EqusSipXbos, Publisher_EqusSipXcis, Publisher_EqusSipTxse, Publisher_EqusSip24Eq, Publisher_EqusSipEprl, Publisher_EqusSipXisx, Publisher_EqusSipEdga, Publisher_EqusSipEdgx, Publisher_EqusSipLtse, Publisher_EqusSipXchi, Publisher_EqusSipXnys, Publisher_EqusSipArcx, Publisher_EqusSipXnas, Publisher_EqusSipMemx, Publisher_EqusSipIexg, Publisher_EqusSipXpsx, Publisher_EqusSipBaty, Publisher_EqusSipBats, Publisher_EqusSipFinn, Publisher_EqusSipFiny, Publisher_EqusSipFinc, Publisher_EqusSipXadf, Publisher_EqusSipCta, Publisher_EqusSipUtp}
	case Dataset_MsciCgif:
		return []Publisher{Publisher_MsciCgifMsci}
	case Dataset_FtseCgif:
		return []Publisher{Publisher_FtseCgifFtse}
	case Dataset_InavCgif:
		return []Publisher{Publisher_InavCgifInav}
	case Dataset_MstarCgif:
		return []Publisher{Publisher_MstarCgifMstar}
	case Dataset_CccyCgif:
		return []Publisher{Publisher_CccyCgifCccy}
	case Dataset_CgiCgif:
		return []Publisher{Publisher_CgiCgifCgi}
	case Dataset_XtksFlex:
		return []Publisher{Publisher_XtksFlexXtks}
	case Dataset_XtktItch:
		return []Publisher{Publisher_XtktItchXtkt}
	case Dataset_XoseItch:
		return []Publisher{Publisher_XoseItchXose}
	case Dataset_DbixCore:
		return []Publisher{Publisher_DbixCoreMain, Publisher_DbixCoreMsci, Publisher_DbixCoreFtse, Publisher_DbixCoreInav, Publisher_DbixCoreMstar, Publisher_DbixCoreCccy, Publisher_DbixCoreCgi, Publisher_DbixCoreCdef}
	default:
		return nil
	}
}

// DatasetFromString converts a string to a Dataset.
// Returns an error if the string is unknown.
func DatasetFromString(str string) (Dataset, error) {
	str = strings.ToUpper(str)
	switch str {
	case "GLBX.MDP3":
		return Dataset_GlbxMdp3, nil
	case "XNAS.ITCH":
		return Dataset_XnasItch, nil
	case "XBOS.ITCH":
		return Dataset_XbosItch, nil
	case "XPSX.ITCH":
		return Dataset_XpsxItch, nil
	case "BATS.PITCH":
		return Dataset_BatsPitch, nil
	case "BATY.PITCH":
		return Dataset_BatyPitch, nil
	case "EDGA.PITCH":
		return Dataset_EdgaPitch, nil
	case "EDGX.PITCH":
		return Dataset_EdgxPitch, nil
	case "XNYS.PILLAR":
		return Dataset_XnysPillar, nil
	case "XCIS.PILLAR":
		return Dataset_XcisPillar, nil
	case "XASE.PILLAR":
		return Dataset_XasePillar, nil
	case "XCHI.PILLAR":
		return Dataset_XchiPillar, nil
	case "XCIS.BBO":
		return Dataset_XcisBbo, nil
	case "XCIS.TRADES":
		return Dataset_XcisTrades, nil
	case "MEMX.MEMOIR":
		return Dataset_MemxMemoir, nil
	case "EPRL.DOM":
		return Dataset_EprlDom, nil
	case "FINN.NLS":
		return Dataset_FinnNls, nil
	case "FINY.TRADES":
		return Dataset_FinyTrades, nil
	case "OPRA.PILLAR":
		return Dataset_OpraPillar, nil
	case "DBEQ.BASIC":
		return Dataset_DbeqBasic, nil
	case "ARCX.PILLAR":
		return Dataset_ArcxPillar, nil
	case "IEXG.TOPS":
		return Dataset_IexgTops, nil
	case "EQUS.PLUS":
		return Dataset_EqusPlus, nil
	case "XNYS.BBO":
		return Dataset_XnysBbo, nil
	case "XNYS.TRADES":
		return Dataset_XnysTrades, nil
	case "XNAS.QBBO":
		return Dataset_XnasQbbo, nil
	case "XNAS.NLS":
		return Dataset_XnasNls, nil
	case "IFEU.IMPACT":
		return Dataset_IfeuImpact, nil
	case "NDEX.IMPACT":
		return Dataset_NdexImpact, nil
	case "EQUS.ALL":
		return Dataset_EqusAll, nil
	case "XNAS.BASIC":
		return Dataset_XnasBasic, nil
	case "EQUS.SUMMARY":
		return Dataset_EqusSummary, nil
	case "XCIS.TRADESBBO":
		return Dataset_XcisTradesbbo, nil
	case "XNYS.TRADESBBO":
		return Dataset_XnysTradesbbo, nil
	case "EQUS.MINI":
		return Dataset_EqusMini, nil
	case "IFUS.IMPACT":
		return Dataset_IfusImpact, nil
	case "IFLL.IMPACT":
		return Dataset_IfllImpact, nil
	case "XEUR.EOBI":
		return Dataset_XeurEobi, nil
	case "XEEE.EOBI":
		return Dataset_XeeeEobi, nil
	case "XCBF.PITCH":
		return Dataset_XcbfPitch, nil
	case "OCEA.MEMOIR":
		return Dataset_OceaMemoir, nil
	case "MAIN.CGIF":
		return Dataset_MainCgif, nil
	case "EQUS.SIP":
		return Dataset_EqusSip, nil
	case "MSCI.CGIF":
		return Dataset_MsciCgif, nil
	case "FTSE.CGIF":
		return Dataset_FtseCgif, nil
	case "INAV.CGIF":
		return Dataset_InavCgif, nil
	case "MSTAR.CGIF":
		return Dataset_MstarCgif, nil
	case "CCCY.CGIF":
		return Dataset_CccyCgif, nil
	case "CGI.CGIF":
		return Dataset_CgiCgif, nil
	case "XTKS.FLEX":
		return Dataset_XtksFlex, nil
	case "XTKT.ITCH":
		return Dataset_XtktItch, nil
	case "XOSE.ITCH":
		return Dataset_XoseItch, nil
	case "DBIX.CORE":
		return Dataset_DbixCore, nil
	default:
		return Dataset_GlbxMdp3, fmt.Errorf("unknown dataset: '%s'", str)
	}
}

func (d Dataset) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d *Dataset) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	ds, err := DatasetFromString(str)
	if err != nil {
		return err
	}
	*d = ds
	return nil
}

// Type implements pflag.Value.Type. Returns "dbn.Dataset".
func (*Dataset) Type() string {
	return "dbn.Dataset"
}

// Set implements the flag.Value interface.
func (d *Dataset) Set(value string) error {
	ds, err := DatasetFromString(value)
	if err == nil {
		*d = ds
	}
	return err
}

///////////////////////////////////////////////////////////////////////////////
// Publisher
///////////////////////////////////////////////////////////////////////////////

// Publisher is a specific Venue from a specific data source.
type Publisher uint16

const (
	// CME Globex MDP 3.0
	Publisher_GlbxMdp3Glbx Publisher = 1
	// Nasdaq TotalView-ITCH
	Publisher_XnasItchXnas Publisher = 2
	// Nasdaq Texas TotalView-ITCH
	Publisher_XbosItchXbos Publisher = 3
	// Nasdaq PSX TotalView-ITCH
	Publisher_XpsxItchXpsx Publisher = 4
	// Cboe BZX Depth
	Publisher_BatsPitchBats Publisher = 5
	// Cboe BYX Depth
	Publisher_BatyPitchBaty Publisher = 6
	// Cboe EDGA Depth
	Publisher_EdgaPitchEdga Publisher = 7
	// Cboe EDGX Depth
	Publisher_EdgxPitchEdgx Publisher = 8
	// NYSE Integrated
	Publisher_XnysPillarXnys Publisher = 9
	// NYSE National Integrated
	Publisher_XcisPillarXcis Publisher = 10
	// NYSE American Integrated
	Publisher_XasePillarXase Publisher = 11
	// NYSE Texas Integrated
	Publisher_XchiPillarXchi Publisher = 12
	// NYSE National BBO
	Publisher_XcisBboXcis Publisher = 13
	// NYSE National Trades
	Publisher_XcisTradesXcis Publisher = 14
	// MEMX MEMOIR Depth
	Publisher_MemxMemoirMemx Publisher = 15
	// MIAX Pearl Depth
	Publisher_EprlDomEprl Publisher = 16
	// FINRA/Nasdaq TRF Carteret
	Publisher_XnasNlsFinn Publisher = 17
	// FINRA/Nasdaq TRF Chicago
	Publisher_XnasNlsFinc Publisher = 18
	// FINRA/NYSE TRF
	Publisher_XnysTradesFiny Publisher = 19
	// OPRA - NYSE American Options
	Publisher_OpraPillarAmxo Publisher = 20
	// OPRA - BOX Options
	Publisher_OpraPillarXbox Publisher = 21
	// OPRA - Cboe Options
	Publisher_OpraPillarXcbo Publisher = 22
	// OPRA - MIAX Emerald
	Publisher_OpraPillarEmld Publisher = 23
	// OPRA - Cboe EDGX Options
	Publisher_OpraPillarEdgo Publisher = 24
	// OPRA - Nasdaq GEMX
	Publisher_OpraPillarGmni Publisher = 25
	// OPRA - Nasdaq ISE
	Publisher_OpraPillarXisx Publisher = 26
	// OPRA - Nasdaq MRX
	Publisher_OpraPillarMcry Publisher = 27
	// OPRA - MIAX Options
	Publisher_OpraPillarXmio Publisher = 28
	// OPRA - NYSE Arca Options
	Publisher_OpraPillarArco Publisher = 29
	// OPRA - Options Price Reporting Authority
	Publisher_OpraPillarOpra Publisher = 30
	// OPRA - MIAX Pearl
	Publisher_OpraPillarMprl Publisher = 31
	// OPRA - Nasdaq Options
	Publisher_OpraPillarXndq Publisher = 32
	// OPRA - Nasdaq Texas Options
	Publisher_OpraPillarXbxo Publisher = 33
	// OPRA - Cboe C2 Options
	Publisher_OpraPillarC2Ox Publisher = 34
	// OPRA - Nasdaq PHLX
	Publisher_OpraPillarXphl Publisher = 35
	// OPRA - Cboe BZX Options
	Publisher_OpraPillarBato Publisher = 36
	// OPRA - MEMX Options
	Publisher_OpraPillarMxop Publisher = 37
	// IEX TOPS
	Publisher_IexgTopsIexg Publisher = 38
	// DBEQ Basic - NYSE Texas
	Publisher_DbeqBasicXchi Publisher = 39
	// DBEQ Basic - NYSE National
	Publisher_DbeqBasicXcis Publisher = 40
	// DBEQ Basic - IEX
	Publisher_DbeqBasicIexg Publisher = 41
	// DBEQ Basic - MIAX Pearl
	Publisher_DbeqBasicEprl Publisher = 42
	// NYSE Arca Integrated
	Publisher_ArcxPillarArcx Publisher = 43
	// NYSE BBO
	Publisher_XnysBboXnys Publisher = 44
	// NYSE Trades
	Publisher_XnysTradesXnys Publisher = 45
	// Nasdaq QBBO
	Publisher_XnasQbboXnas Publisher = 46
	// Nasdaq Trades
	Publisher_XnasNlsXnas Publisher = 47
	// Databento US Equities Plus - NYSE Texas
	Publisher_EqusPlusXchi Publisher = 48
	// Databento US Equities Plus - NYSE National
	Publisher_EqusPlusXcis Publisher = 49
	// Databento US Equities Plus - IEX
	Publisher_EqusPlusIexg Publisher = 50
	// Databento US Equities Plus - MIAX Pearl
	Publisher_EqusPlusEprl Publisher = 51
	// Databento US Equities Plus - Nasdaq
	Publisher_EqusPlusXnas Publisher = 52
	// Databento US Equities Plus - NYSE
	Publisher_EqusPlusXnys Publisher = 53
	// Databento US Equities Plus - FINRA/Nasdaq TRF Carteret
	Publisher_EqusPlusFinn Publisher = 54
	// Databento US Equities Plus - FINRA/NYSE TRF
	Publisher_EqusPlusFiny Publisher = 55
	// Databento US Equities Plus - FINRA/Nasdaq TRF Chicago
	Publisher_EqusPlusFinc Publisher = 56
	// ICE Europe Commodities
	Publisher_IfeuImpactIfeu Publisher = 57
	// ICE Endex
	Publisher_NdexImpactNdex Publisher = 58
	// Databento US Equities Basic - Consolidated
	Publisher_DbeqBasicDbeq Publisher = 59
	// EQUS Plus - Consolidated
	Publisher_EqusPlusEqus Publisher = 60
	// OPRA - MIAX Sapphire
	Publisher_OpraPillarSphr Publisher = 61
	// Databento US Equities (All Feeds) - NYSE Texas
	Publisher_EqusAllXchi Publisher = 62
	// Databento US Equities (All Feeds) - NYSE National
	Publisher_EqusAllXcis Publisher = 63
	// Databento US Equities (All Feeds) - IEX
	Publisher_EqusAllIexg Publisher = 64
	// Databento US Equities (All Feeds) - MIAX Pearl
	Publisher_EqusAllEprl Publisher = 65
	// Databento US Equities (All Feeds) - Nasdaq
	Publisher_EqusAllXnas Publisher = 66
	// Databento US Equities (All Feeds) - NYSE
	Publisher_EqusAllXnys Publisher = 67
	// Databento US Equities (All Feeds) - FINRA/Nasdaq TRF Carteret
	Publisher_EqusAllFinn Publisher = 68
	// Databento US Equities (All Feeds) - FINRA/NYSE TRF
	Publisher_EqusAllFiny Publisher = 69
	// Databento US Equities (All Feeds) - FINRA/Nasdaq TRF Chicago
	Publisher_EqusAllFinc Publisher = 70
	// Databento US Equities (All Feeds) - Cboe BZX
	Publisher_EqusAllBats Publisher = 71
	// Databento US Equities (All Feeds) - Cboe BYX
	Publisher_EqusAllBaty Publisher = 72
	// Databento US Equities (All Feeds) - Cboe EDGA
	Publisher_EqusAllEdga Publisher = 73
	// Databento US Equities (All Feeds) - Cboe EDGX
	Publisher_EqusAllEdgx Publisher = 74
	// Databento US Equities (All Feeds) - Nasdaq Texas
	Publisher_EqusAllXbos Publisher = 75
	// Databento US Equities (All Feeds) - Nasdaq PSX
	Publisher_EqusAllXpsx Publisher = 76
	// Databento US Equities (All Feeds) - MEMX
	Publisher_EqusAllMemx Publisher = 77
	// Databento US Equities (All Feeds) - NYSE American
	Publisher_EqusAllXase Publisher = 78
	// Databento US Equities (All Feeds) - NYSE Arca
	Publisher_EqusAllArcx Publisher = 79
	// Databento US Equities (All Feeds) - Long-Term Stock Exchange
	Publisher_EqusAllLtse Publisher = 80
	// Nasdaq Basic - Nasdaq
	Publisher_XnasBasicXnas Publisher = 81
	// Nasdaq Basic - FINRA/Nasdaq TRF Carteret
	Publisher_XnasBasicFinn Publisher = 82
	// Nasdaq Basic - FINRA/Nasdaq TRF Chicago
	Publisher_XnasBasicFinc Publisher = 83
	// ICE Europe - Off-Market Trades
	Publisher_IfeuImpactXoff Publisher = 84
	// ICE Endex - Off-Market Trades
	Publisher_NdexImpactXoff Publisher = 85
	// Nasdaq NLS - Nasdaq Texas
	Publisher_XnasNlsXbos Publisher = 86
	// Nasdaq NLS - Nasdaq PSX
	Publisher_XnasNlsXpsx Publisher = 87
	// Nasdaq Basic - Nasdaq Texas
	Publisher_XnasBasicXbos Publisher = 88
	// Nasdaq Basic - Nasdaq PSX
	Publisher_XnasBasicXpsx Publisher = 89
	// Databento Equities Summary
	Publisher_EqusSummaryEqus Publisher = 90
	// NYSE National Trades and BBO
	Publisher_XcisTradesbboXcis Publisher = 91
	// NYSE Trades and BBO
	Publisher_XnysTradesbboXnys Publisher = 92
	// Nasdaq Basic - Consolidated
	Publisher_XnasBasicEqus Publisher = 93
	// Databento US Equities (All Feeds) - Consolidated
	Publisher_EqusAllEqus Publisher = 94
	// Databento US Equities Mini
	Publisher_EqusMiniEqus Publisher = 95
	// NYSE Trades - Consolidated
	Publisher_XnysTradesEqus Publisher = 96
	// ICE Futures US
	Publisher_IfusImpactIfus Publisher = 97
	// ICE Futures US - Off-Market Trades
	Publisher_IfusImpactXoff Publisher = 98
	// ICE Europe Financials
	Publisher_IfllImpactIfll Publisher = 99
	// ICE Europe Financials - Off-Market Trades
	Publisher_IfllImpactXoff Publisher = 100
	// Eurex EOBI
	Publisher_XeurEobiXeur Publisher = 101
	// European Energy Exchange EOBI
	Publisher_XeeeEobiXeee Publisher = 102
	// Eurex EOBI - Off-Market Trades
	Publisher_XeurEobiXoff Publisher = 103
	// European Energy Exchange EOBI - Off-Market Trades
	Publisher_XeeeEobiXoff Publisher = 104
	// Cboe Futures Exchange (CFE)
	Publisher_XcbfPitchXcbf Publisher = 105
	// Cboe Futures Exchange (CFE) - Off-Market Trades
	Publisher_XcbfPitchXoff Publisher = 106
	// Blue Ocean ATS MEMOIR
	Publisher_OceaMemoirOcea Publisher = 107
	// OPRA - MEMX MX2 Options
	Publisher_OpraPillarMxto Publisher = 108
	// OPRA - IEX Options LLC
	Publisher_OpraPillarIexo Publisher = 109
	// Cboe Global Indices Feed
	Publisher_MainCgifMain Publisher = 110
	// US Equities SIP - NYSE American
	Publisher_EqusSipXase Publisher = 111
	// US Equities SIP - Nasdaq Texas
	Publisher_EqusSipXbos Publisher = 112
	// US Equities SIP - NYSE National
	Publisher_EqusSipXcis Publisher = 113
	// US Equities SIP - Texas Stock Exchange
	Publisher_EqusSipTxse Publisher = 114
	// US Equities SIP - 24X National Exchange
	Publisher_EqusSip24Eq Publisher = 115
	// US Equities SIP - MIAX Pearl
	Publisher_EqusSipEprl Publisher = 116
	// US Equities SIP - Nasdaq ISE
	Publisher_EqusSipXisx Publisher = 117
	// US Equities SIP - Cboe EDGA
	Publisher_EqusSipEdga Publisher = 118
	// US Equities SIP - Cboe EDGX
	Publisher_EqusSipEdgx Publisher = 119
	// US Equities SIP - Long-Term Stock Exchange
	Publisher_EqusSipLtse Publisher = 120
	// US Equities SIP - NYSE Texas
	Publisher_EqusSipXchi Publisher = 121
	// US Equities SIP - NYSE
	Publisher_EqusSipXnys Publisher = 122
	// US Equities SIP - NYSE Arca
	Publisher_EqusSipArcx Publisher = 123
	// US Equities SIP - Nasdaq
	Publisher_EqusSipXnas Publisher = 124
	// US Equities SIP - MEMX
	Publisher_EqusSipMemx Publisher = 125
	// US Equities SIP - IEX
	Publisher_EqusSipIexg Publisher = 126
	// US Equities SIP - Nasdaq PSX
	Publisher_EqusSipXpsx Publisher = 127
	// US Equities SIP - Cboe BYX
	Publisher_EqusSipBaty Publisher = 128
	// US Equities SIP - Cboe BZX
	Publisher_EqusSipBats Publisher = 129
	// US Equities SIP - FINRA/Nasdaq TRF Carteret
	Publisher_EqusSipFinn Publisher = 130
	// US Equities SIP - FINRA/NYSE TRF
	Publisher_EqusSipFiny Publisher = 131
	// US Equities SIP - FINRA/Nasdaq TRF Chicago
	Publisher_EqusSipFinc Publisher = 132
	// US Equities SIP - FINRA Alternative Display Facility
	Publisher_EqusSipXadf Publisher = 133
	// US Equities SIP - CTA
	Publisher_EqusSipCta Publisher = 134
	// US Equities SIP - UTP
	Publisher_EqusSipUtp Publisher = 135
	// CGIF - Morgan Stanley Capital International
	Publisher_MsciCgifMsci Publisher = 136
	// CGIF - FTSE Russell
	Publisher_FtseCgifFtse Publisher = 137
	// CGIF - Intraday Net Asset Values
	Publisher_InavCgifInav Publisher = 138
	// CGIF - Morning Star
	Publisher_MstarCgifMstar Publisher = 139
	// CGIF - Crypto Currency
	Publisher_CccyCgifCccy Publisher = 140
	// CGIF - CBOE Global Indices
	Publisher_CgiCgifCgi Publisher = 141
	// CGIF - Definitions
	Publisher_MainCgifCdef Publisher = 142
	// JPX - Tokyo Stock Exchange
	Publisher_XtksFlexXtks Publisher = 143
	// JPX - Tokyo Commodity Exchange
	Publisher_XtktItchXtkt Publisher = 144
	// JPX - Osaka Exchange
	Publisher_XoseItchXose Publisher = 145
	// CGIF - Main Feed
	Publisher_DbixCoreMain Publisher = 146
	// CGIF - Morgan Stanley Capital International
	Publisher_DbixCoreMsci Publisher = 147
	// CGIF - FTSE Russell
	Publisher_DbixCoreFtse Publisher = 148
	// CGIF - Intraday Net Asset Values
	Publisher_DbixCoreInav Publisher = 149
	// CGIF - Morning Star
	Publisher_DbixCoreMstar Publisher = 150
	// CGIF - Crypto Currency
	Publisher_DbixCoreCccy Publisher = 151
	// CGIF - CBOE Global Indices
	Publisher_DbixCoreCgi Publisher = 152
	// CGIF - Definitions
	Publisher_DbixCoreCdef Publisher = 153
)

const PUBLISHER_COUNT = 153

// Returns the string representation of the Publisher, or empty string if unknown.
func (p Publisher) String() string {
	switch p {
	case Publisher_GlbxMdp3Glbx:
		return "GLBX.MDP3.GLBX"
	case Publisher_XnasItchXnas:
		return "XNAS.ITCH.XNAS"
	case Publisher_XbosItchXbos:
		return "XBOS.ITCH.XBOS"
	case Publisher_XpsxItchXpsx:
		return "XPSX.ITCH.XPSX"
	case Publisher_BatsPitchBats:
		return "BATS.PITCH.BATS"
	case Publisher_BatyPitchBaty:
		return "BATY.PITCH.BATY"
	case Publisher_EdgaPitchEdga:
		return "EDGA.PITCH.EDGA"
	case Publisher_EdgxPitchEdgx:
		return "EDGX.PITCH.EDGX"
	case Publisher_XnysPillarXnys:
		return "XNYS.PILLAR.XNYS"
	case Publisher_XcisPillarXcis:
		return "XCIS.PILLAR.XCIS"
	case Publisher_XasePillarXase:
		return "XASE.PILLAR.XASE"
	case Publisher_XchiPillarXchi:
		return "XCHI.PILLAR.XCHI"
	case Publisher_XcisBboXcis:
		return "XCIS.BBO.XCIS"
	case Publisher_XcisTradesXcis:
		return "XCIS.TRADES.XCIS"
	case Publisher_MemxMemoirMemx:
		return "MEMX.MEMOIR.MEMX"
	case Publisher_EprlDomEprl:
		return "EPRL.DOM.EPRL"
	case Publisher_XnasNlsFinn:
		return "XNAS.NLS.FINN"
	case Publisher_XnasNlsFinc:
		return "XNAS.NLS.FINC"
	case Publisher_XnysTradesFiny:
		return "XNYS.TRADES.FINY"
	case Publisher_OpraPillarAmxo:
		return "OPRA.PILLAR.AMXO"
	case Publisher_OpraPillarXbox:
		return "OPRA.PILLAR.XBOX"
	case Publisher_OpraPillarXcbo:
		return "OPRA.PILLAR.XCBO"
	case Publisher_OpraPillarEmld:
		return "OPRA.PILLAR.EMLD"
	case Publisher_OpraPillarEdgo:
		return "OPRA.PILLAR.EDGO"
	case Publisher_OpraPillarGmni:
		return "OPRA.PILLAR.GMNI"
	case Publisher_OpraPillarXisx:
		return "OPRA.PILLAR.XISX"
	case Publisher_OpraPillarMcry:
		return "OPRA.PILLAR.MCRY"
	case Publisher_OpraPillarXmio:
		return "OPRA.PILLAR.XMIO"
	case Publisher_OpraPillarArco:
		return "OPRA.PILLAR.ARCO"
	case Publisher_OpraPillarOpra:
		return "OPRA.PILLAR.OPRA"
	case Publisher_OpraPillarMprl:
		return "OPRA.PILLAR.MPRL"
	case Publisher_OpraPillarXndq:
		return "OPRA.PILLAR.XNDQ"
	case Publisher_OpraPillarXbxo:
		return "OPRA.PILLAR.XBXO"
	case Publisher_OpraPillarC2Ox:
		return "OPRA.PILLAR.C2OX"
	case Publisher_OpraPillarXphl:
		return "OPRA.PILLAR.XPHL"
	case Publisher_OpraPillarBato:
		return "OPRA.PILLAR.BATO"
	case Publisher_OpraPillarMxop:
		return "OPRA.PILLAR.MXOP"
	case Publisher_IexgTopsIexg:
		return "IEXG.TOPS.IEXG"
	case Publisher_DbeqBasicXchi:
		return "DBEQ.BASIC.XCHI"
	case Publisher_DbeqBasicXcis:
		return "DBEQ.BASIC.XCIS"
	case Publisher_DbeqBasicIexg:
		return "DBEQ.BASIC.IEXG"
	case Publisher_DbeqBasicEprl:
		return "DBEQ.BASIC.EPRL"
	case Publisher_ArcxPillarArcx:
		return "ARCX.PILLAR.ARCX"
	case Publisher_XnysBboXnys:
		return "XNYS.BBO.XNYS"
	case Publisher_XnysTradesXnys:
		return "XNYS.TRADES.XNYS"
	case Publisher_XnasQbboXnas:
		return "XNAS.QBBO.XNAS"
	case Publisher_XnasNlsXnas:
		return "XNAS.NLS.XNAS"
	case Publisher_EqusPlusXchi:
		return "EQUS.PLUS.XCHI"
	case Publisher_EqusPlusXcis:
		return "EQUS.PLUS.XCIS"
	case Publisher_EqusPlusIexg:
		return "EQUS.PLUS.IEXG"
	case Publisher_EqusPlusEprl:
		return "EQUS.PLUS.EPRL"
	case Publisher_EqusPlusXnas:
		return "EQUS.PLUS.XNAS"
	case Publisher_EqusPlusXnys:
		return "EQUS.PLUS.XNYS"
	case Publisher_EqusPlusFinn:
		return "EQUS.PLUS.FINN"
	case Publisher_EqusPlusFiny:
		return "EQUS.PLUS.FINY"
	case Publisher_EqusPlusFinc:
		return "EQUS.PLUS.FINC"
	case Publisher_IfeuImpactIfeu:
		return "IFEU.IMPACT.IFEU"
	case Publisher_NdexImpactNdex:
		return "NDEX.IMPACT.NDEX"
	case Publisher_DbeqBasicDbeq:
		return "DBEQ.BASIC.DBEQ"
	case Publisher_EqusPlusEqus:
		return "EQUS.PLUS.EQUS"
	case Publisher_OpraPillarSphr:
		return "OPRA.PILLAR.SPHR"
	case Publisher_EqusAllXchi:
		return "EQUS.ALL.XCHI"
	case Publisher_EqusAllXcis:
		return "EQUS.ALL.XCIS"
	case Publisher_EqusAllIexg:
		return "EQUS.ALL.IEXG"
	case Publisher_EqusAllEprl:
		return "EQUS.ALL.EPRL"
	case Publisher_EqusAllXnas:
		return "EQUS.ALL.XNAS"
	case Publisher_EqusAllXnys:
		return "EQUS.ALL.XNYS"
	case Publisher_EqusAllFinn:
		return "EQUS.ALL.FINN"
	case Publisher_EqusAllFiny:
		return "EQUS.ALL.FINY"
	case Publisher_EqusAllFinc:
		return "EQUS.ALL.FINC"
	case Publisher_EqusAllBats:
		return "EQUS.ALL.BATS"
	case Publisher_EqusAllBaty:
		return "EQUS.ALL.BATY"
	case Publisher_EqusAllEdga:
		return "EQUS.ALL.EDGA"
	case Publisher_EqusAllEdgx:
		return "EQUS.ALL.EDGX"
	case Publisher_EqusAllXbos:
		return "EQUS.ALL.XBOS"
	case Publisher_EqusAllXpsx:
		return "EQUS.ALL.XPSX"
	case Publisher_EqusAllMemx:
		return "EQUS.ALL.MEMX"
	case Publisher_EqusAllXase:
		return "EQUS.ALL.XASE"
	case Publisher_EqusAllArcx:
		return "EQUS.ALL.ARCX"
	case Publisher_EqusAllLtse:
		return "EQUS.ALL.LTSE"
	case Publisher_XnasBasicXnas:
		return "XNAS.BASIC.XNAS"
	case Publisher_XnasBasicFinn:
		return "XNAS.BASIC.FINN"
	case Publisher_XnasBasicFinc:
		return "XNAS.BASIC.FINC"
	case Publisher_IfeuImpactXoff:
		return "IFEU.IMPACT.XOFF"
	case Publisher_NdexImpactXoff:
		return "NDEX.IMPACT.XOFF"
	case Publisher_XnasNlsXbos:
		return "XNAS.NLS.XBOS"
	case Publisher_XnasNlsXpsx:
		return "XNAS.NLS.XPSX"
	case Publisher_XnasBasicXbos:
		return "XNAS.BASIC.XBOS"
	case Publisher_XnasBasicXpsx:
		return "XNAS.BASIC.XPSX"
	case Publisher_EqusSummaryEqus:
		return "EQUS.SUMMARY.EQUS"
	case Publisher_XcisTradesbboXcis:
		return "XCIS.TRADESBBO.XCIS"
	case Publisher_XnysTradesbboXnys:
		return "XNYS.TRADESBBO.XNYS"
	case Publisher_XnasBasicEqus:
		return "XNAS.BASIC.EQUS"
	case Publisher_EqusAllEqus:
		return "EQUS.ALL.EQUS"
	case Publisher_EqusMiniEqus:
		return "EQUS.MINI.EQUS"
	case Publisher_XnysTradesEqus:
		return "XNYS.TRADES.EQUS"
	case Publisher_IfusImpactIfus:
		return "IFUS.IMPACT.IFUS"
	case Publisher_IfusImpactXoff:
		return "IFUS.IMPACT.XOFF"
	case Publisher_IfllImpactIfll:
		return "IFLL.IMPACT.IFLL"
	case Publisher_IfllImpactXoff:
		return "IFLL.IMPACT.XOFF"
	case Publisher_XeurEobiXeur:
		return "XEUR.EOBI.XEUR"
	case Publisher_XeeeEobiXeee:
		return "XEEE.EOBI.XEEE"
	case Publisher_XeurEobiXoff:
		return "XEUR.EOBI.XOFF"
	case Publisher_XeeeEobiXoff:
		return "XEEE.EOBI.XOFF"
	case Publisher_XcbfPitchXcbf:
		return "XCBF.PITCH.XCBF"
	case Publisher_XcbfPitchXoff:
		return "XCBF.PITCH.XOFF"
	case Publisher_OceaMemoirOcea:
		return "OCEA.MEMOIR.OCEA"
	case Publisher_OpraPillarMxto:
		return "OPRA.PILLAR.MXTO"
	case Publisher_OpraPillarIexo:
		return "OPRA.PILLAR.IEXO"
	case Publisher_MainCgifMain:
		return "MAIN.CGIF.MAIN"
	case Publisher_EqusSipXase:
		return "EQUS.SIP.XASE"
	case Publisher_EqusSipXbos:
		return "EQUS.SIP.XBOS"
	case Publisher_EqusSipXcis:
		return "EQUS.SIP.XCIS"
	case Publisher_EqusSipTxse:
		return "EQUS.SIP.TXSE"
	case Publisher_EqusSip24Eq:
		return "EQUS.SIP.24EQ"
	case Publisher_EqusSipEprl:
		return "EQUS.SIP.EPRL"
	case Publisher_EqusSipXisx:
		return "EQUS.SIP.XISX"
	case Publisher_EqusSipEdga:
		return "EQUS.SIP.EDGA"
	case Publisher_EqusSipEdgx:
		return "EQUS.SIP.EDGX"
	case Publisher_EqusSipLtse:
		return "EQUS.SIP.LTSE"
	case Publisher_EqusSipXchi:
		return "EQUS.SIP.XCHI"
	case Publisher_EqusSipXnys:
		return "EQUS.SIP.XNYS"
	case Publisher_EqusSipArcx:
		return "EQUS.SIP.ARCX"
	case Publisher_EqusSipXnas:
		return "EQUS.SIP.XNAS"
	case Publisher_EqusSipMemx:
		return "EQUS.SIP.MEMX"
	case Publisher_EqusSipIexg:
		return "EQUS.SIP.IEXG"
	case Publisher_EqusSipXpsx:
		return "EQUS.SIP.XPSX"
	case Publisher_EqusSipBaty:
		return "EQUS.SIP.BATY"
	case Publisher_EqusSipBats:
		return "EQUS.SIP.BATS"
	case Publisher_EqusSipFinn:
		return "EQUS.SIP.FINN"
	case Publisher_EqusSipFiny:
		return "EQUS.SIP.FINY"
	case Publisher_EqusSipFinc:
		return "EQUS.SIP.FINC"
	case Publisher_EqusSipXadf:
		return "EQUS.SIP.XADF"
	case Publisher_EqusSipCta:
		return "EQUS.SIP.CTA"
	case Publisher_EqusSipUtp:
		return "EQUS.SIP.UTP"
	case Publisher_MsciCgifMsci:
		return "MSCI.CGIF.MSCI"
	case Publisher_FtseCgifFtse:
		return "FTSE.CGIF.FTSE"
	case Publisher_InavCgifInav:
		return "INAV.CGIF.INAV"
	case Publisher_MstarCgifMstar:
		return "MSTAR.CGIF.MSTAR"
	case Publisher_CccyCgifCccy:
		return "CCCY.CGIF.CCCY"
	case Publisher_CgiCgifCgi:
		return "CGI.CGIF.CGI"
	case Publisher_MainCgifCdef:
		return "MAIN.CGIF.CDEF"
	case Publisher_XtksFlexXtks:
		return "XTKS.FLEX.XTKS"
	case Publisher_XtktItchXtkt:
		return "XTKT.ITCH.XTKT"
	case Publisher_XoseItchXose:
		return "XOSE.ITCH.XOSE"
	case Publisher_DbixCoreMain:
		return "DBIX.CORE.MAIN"
	case Publisher_DbixCoreMsci:
		return "DBIX.CORE.MSCI"
	case Publisher_DbixCoreFtse:
		return "DBIX.CORE.FTSE"
	case Publisher_DbixCoreInav:
		return "DBIX.CORE.INAV"
	case Publisher_DbixCoreMstar:
		return "DBIX.CORE.MSTAR"
	case Publisher_DbixCoreCccy:
		return "DBIX.CORE.CCCY"
	case Publisher_DbixCoreCgi:
		return "DBIX.CORE.CGI"
	case Publisher_DbixCoreCdef:
		return "DBIX.CORE.CDEF"
	default:
		return ""
	}
}

// Venue returns the publisher's Venue.
func (p Publisher) Venue() Venue {
	switch p {
	case Publisher_GlbxMdp3Glbx:
		return Venue_Glbx
	case Publisher_XnasItchXnas:
		return Venue_Xnas
	case Publisher_XbosItchXbos:
		return Venue_Xbos
	case Publisher_XpsxItchXpsx:
		return Venue_Xpsx
	case Publisher_BatsPitchBats:
		return Venue_Bats
	case Publisher_BatyPitchBaty:
		return Venue_Baty
	case Publisher_EdgaPitchEdga:
		return Venue_Edga
	case Publisher_EdgxPitchEdgx:
		return Venue_Edgx
	case Publisher_XnysPillarXnys:
		return Venue_Xnys
	case Publisher_XcisPillarXcis:
		return Venue_Xcis
	case Publisher_XasePillarXase:
		return Venue_Xase
	case Publisher_XchiPillarXchi:
		return Venue_Xchi
	case Publisher_XcisBboXcis:
		return Venue_Xcis
	case Publisher_XcisTradesXcis:
		return Venue_Xcis
	case Publisher_MemxMemoirMemx:
		return Venue_Memx
	case Publisher_EprlDomEprl:
		return Venue_Eprl
	case Publisher_XnasNlsFinn:
		return Venue_Finn
	case Publisher_XnasNlsFinc:
		return Venue_Finc
	case Publisher_XnysTradesFiny:
		return Venue_Finy
	case Publisher_OpraPillarAmxo:
		return Venue_Amxo
	case Publisher_OpraPillarXbox:
		return Venue_Xbox
	case Publisher_OpraPillarXcbo:
		return Venue_Xcbo
	case Publisher_OpraPillarEmld:
		return Venue_Emld
	case Publisher_OpraPillarEdgo:
		return Venue_Edgo
	case Publisher_OpraPillarGmni:
		return Venue_Gmni
	case Publisher_OpraPillarXisx:
		return Venue_Xisx
	case Publisher_OpraPillarMcry:
		return Venue_Mcry
	case Publisher_OpraPillarXmio:
		return Venue_Xmio
	case Publisher_OpraPillarArco:
		return Venue_Arco
	case Publisher_OpraPillarOpra:
		return Venue_Opra
	case Publisher_OpraPillarMprl:
		return Venue_Mprl
	case Publisher_OpraPillarXndq:
		return Venue_Xndq
	case Publisher_OpraPillarXbxo:
		return Venue_Xbxo
	case Publisher_OpraPillarC2Ox:
		return Venue_C2Ox
	case Publisher_OpraPillarXphl:
		return Venue_Xphl
	case Publisher_OpraPillarBato:
		return Venue_Bato
	case Publisher_OpraPillarMxop:
		return Venue_Mxop
	case Publisher_IexgTopsIexg:
		return Venue_Iexg
	case Publisher_DbeqBasicXchi:
		return Venue_Xchi
	case Publisher_DbeqBasicXcis:
		return Venue_Xcis
	case Publisher_DbeqBasicIexg:
		return Venue_Iexg
	case Publisher_DbeqBasicEprl:
		return Venue_Eprl
	case Publisher_ArcxPillarArcx:
		return Venue_Arcx
	case Publisher_XnysBboXnys:
		return Venue_Xnys
	case Publisher_XnysTradesXnys:
		return Venue_Xnys
	case Publisher_XnasQbboXnas:
		return Venue_Xnas
	case Publisher_XnasNlsXnas:
		return Venue_Xnas
	case Publisher_EqusPlusXchi:
		return Venue_Xchi
	case Publisher_EqusPlusXcis:
		return Venue_Xcis
	case Publisher_EqusPlusIexg:
		return Venue_Iexg
	case Publisher_EqusPlusEprl:
		return Venue_Eprl
	case Publisher_EqusPlusXnas:
		return Venue_Xnas
	case Publisher_EqusPlusXnys:
		return Venue_Xnys
	case Publisher_EqusPlusFinn:
		return Venue_Finn
	case Publisher_EqusPlusFiny:
		return Venue_Finy
	case Publisher_EqusPlusFinc:
		return Venue_Finc
	case Publisher_IfeuImpactIfeu:
		return Venue_Ifeu
	case Publisher_NdexImpactNdex:
		return Venue_Ndex
	case Publisher_DbeqBasicDbeq:
		return Venue_Dbeq
	case Publisher_EqusPlusEqus:
		return Venue_Equs
	case Publisher_OpraPillarSphr:
		return Venue_Sphr
	case Publisher_EqusAllXchi:
		return Venue_Xchi
	case Publisher_EqusAllXcis:
		return Venue_Xcis
	case Publisher_EqusAllIexg:
		return Venue_Iexg
	case Publisher_EqusAllEprl:
		return Venue_Eprl
	case Publisher_EqusAllXnas:
		return Venue_Xnas
	case Publisher_EqusAllXnys:
		return Venue_Xnys
	case Publisher_EqusAllFinn:
		return Venue_Finn
	case Publisher_EqusAllFiny:
		return Venue_Finy
	case Publisher_EqusAllFinc:
		return Venue_Finc
	case Publisher_EqusAllBats:
		return Venue_Bats
	case Publisher_EqusAllBaty:
		return Venue_Baty
	case Publisher_EqusAllEdga:
		return Venue_Edga
	case Publisher_EqusAllEdgx:
		return Venue_Edgx
	case Publisher_EqusAllXbos:
		return Venue_Xbos
	case Publisher_EqusAllXpsx:
		return Venue_Xpsx
	case Publisher_EqusAllMemx:
		return Venue_Memx
	case Publisher_EqusAllXase:
		return Venue_Xase
	case Publisher_EqusAllArcx:
		return Venue_Arcx
	case Publisher_EqusAllLtse:
		return Venue_Ltse
	case Publisher_XnasBasicXnas:
		return Venue_Xnas
	case Publisher_XnasBasicFinn:
		return Venue_Finn
	case Publisher_XnasBasicFinc:
		return Venue_Finc
	case Publisher_IfeuImpactXoff:
		return Venue_Xoff
	case Publisher_NdexImpactXoff:
		return Venue_Xoff
	case Publisher_XnasNlsXbos:
		return Venue_Xbos
	case Publisher_XnasNlsXpsx:
		return Venue_Xpsx
	case Publisher_XnasBasicXbos:
		return Venue_Xbos
	case Publisher_XnasBasicXpsx:
		return Venue_Xpsx
	case Publisher_EqusSummaryEqus:
		return Venue_Equs
	case Publisher_XcisTradesbboXcis:
		return Venue_Xcis
	case Publisher_XnysTradesbboXnys:
		return Venue_Xnys
	case Publisher_XnasBasicEqus:
		return Venue_Equs
	case Publisher_EqusAllEqus:
		return Venue_Equs
	case Publisher_EqusMiniEqus:
		return Venue_Equs
	case Publisher_XnysTradesEqus:
		return Venue_Equs
	case Publisher_IfusImpactIfus:
		return Venue_Ifus
	case Publisher_IfusImpactXoff:
		return Venue_Xoff
	case Publisher_IfllImpactIfll:
		return Venue_Ifll
	case Publisher_IfllImpactXoff:
		return Venue_Xoff
	case Publisher_XeurEobiXeur:
		return Venue_Xeur
	case Publisher_XeeeEobiXeee:
		return Venue_Xeee
	case Publisher_XeurEobiXoff:
		return Venue_Xoff
	case Publisher_XeeeEobiXoff:
		return Venue_Xoff
	case Publisher_XcbfPitchXcbf:
		return Venue_Xcbf
	case Publisher_XcbfPitchXoff:
		return Venue_Xoff
	case Publisher_OceaMemoirOcea:
		return Venue_Ocea
	case Publisher_OpraPillarMxto:
		return Venue_Mxto
	case Publisher_OpraPillarIexo:
		return Venue_Iexo
	case Publisher_MainCgifMain:
		return Venue_Main
	case Publisher_EqusSipXase:
		return Venue_Xase
	case Publisher_EqusSipXbos:
		return Venue_Xbos
	case Publisher_EqusSipXcis:
		return Venue_Xcis
	case Publisher_EqusSipTxse:
		return Venue_Txse
	case Publisher_EqusSip24Eq:
		return Venue__24Eq
	case Publisher_EqusSipEprl:
		return Venue_Eprl
	case Publisher_EqusSipXisx:
		return Venue_Xisx
	case Publisher_EqusSipEdga:
		return Venue_Edga
	case Publisher_EqusSipEdgx:
		return Venue_Edgx
	case Publisher_EqusSipLtse:
		return Venue_Ltse
	case Publisher_EqusSipXchi:
		return Venue_Xchi
	case Publisher_EqusSipXnys:
		return Venue_Xnys
	case Publisher_EqusSipArcx:
		return Venue_Arcx
	case Publisher_EqusSipXnas:
		return Venue_Xnas
	case Publisher_EqusSipMemx:
		return Venue_Memx
	case Publisher_EqusSipIexg:
		return Venue_Iexg
	case Publisher_EqusSipXpsx:
		return Venue_Xpsx
	case Publisher_EqusSipBaty:
		return Venue_Baty
	case Publisher_EqusSipBats:
		return Venue_Bats
	case Publisher_EqusSipFinn:
		return Venue_Finn
	case Publisher_EqusSipFiny:
		return Venue_Finy
	case Publisher_EqusSipFinc:
		return Venue_Finc
	case Publisher_EqusSipXadf:
		return Venue_Xadf
	case Publisher_EqusSipCta:
		return Venue_Cta
	case Publisher_EqusSipUtp:
		return Venue_Utp
	case Publisher_MsciCgifMsci:
		return Venue_Msci
	case Publisher_FtseCgifFtse:
		return Venue_Ftse
	case Publisher_InavCgifInav:
		return Venue_Inav
	case Publisher_MstarCgifMstar:
		return Venue_Mstar
	case Publisher_CccyCgifCccy:
		return Venue_Cccy
	case Publisher_CgiCgifCgi:
		return Venue_Cgi
	case Publisher_MainCgifCdef:
		return Venue_Cdef
	case Publisher_XtksFlexXtks:
		return Venue_Xtks
	case Publisher_XtktItchXtkt:
		return Venue_Xtkt
	case Publisher_XoseItchXose:
		return Venue_Xose
	case Publisher_DbixCoreMain:
		return Venue_Main
	case Publisher_DbixCoreMsci:
		return Venue_Msci
	case Publisher_DbixCoreFtse:
		return Venue_Ftse
	case Publisher_DbixCoreInav:
		return Venue_Inav
	case Publisher_DbixCoreMstar:
		return Venue_Mstar
	case Publisher_DbixCoreCccy:
		return Venue_Cccy
	case Publisher_DbixCoreCgi:
		return Venue_Cgi
	case Publisher_DbixCoreCdef:
		return Venue_Cdef
	default:
		return 0
	}
}

// Dataset returns the publisher's Dataset.
func (p Publisher) Dataset() Dataset {
	switch p {
	case Publisher_GlbxMdp3Glbx:
		return Dataset_GlbxMdp3
	case Publisher_XnasItchXnas:
		return Dataset_XnasItch
	case Publisher_XbosItchXbos:
		return Dataset_XbosItch
	case Publisher_XpsxItchXpsx:
		return Dataset_XpsxItch
	case Publisher_BatsPitchBats:
		return Dataset_BatsPitch
	case Publisher_BatyPitchBaty:
		return Dataset_BatyPitch
	case Publisher_EdgaPitchEdga:
		return Dataset_EdgaPitch
	case Publisher_EdgxPitchEdgx:
		return Dataset_EdgxPitch
	case Publisher_XnysPillarXnys:
		return Dataset_XnysPillar
	case Publisher_XcisPillarXcis:
		return Dataset_XcisPillar
	case Publisher_XasePillarXase:
		return Dataset_XasePillar
	case Publisher_XchiPillarXchi:
		return Dataset_XchiPillar
	case Publisher_XcisBboXcis:
		return Dataset_XcisBbo
	case Publisher_XcisTradesXcis:
		return Dataset_XcisTrades
	case Publisher_MemxMemoirMemx:
		return Dataset_MemxMemoir
	case Publisher_EprlDomEprl:
		return Dataset_EprlDom
	case Publisher_OpraPillarAmxo, Publisher_OpraPillarXbox, Publisher_OpraPillarXcbo, Publisher_OpraPillarEmld, Publisher_OpraPillarEdgo, Publisher_OpraPillarGmni, Publisher_OpraPillarXisx, Publisher_OpraPillarMcry, Publisher_OpraPillarXmio, Publisher_OpraPillarArco, Publisher_OpraPillarOpra, Publisher_OpraPillarMprl, Publisher_OpraPillarXndq, Publisher_OpraPillarXbxo, Publisher_OpraPillarC2Ox, Publisher_OpraPillarXphl, Publisher_OpraPillarBato, Publisher_OpraPillarMxop, Publisher_OpraPillarSphr, Publisher_OpraPillarMxto, Publisher_OpraPillarIexo:
		return Dataset_OpraPillar
	case Publisher_DbeqBasicXchi, Publisher_DbeqBasicXcis, Publisher_DbeqBasicIexg, Publisher_DbeqBasicEprl, Publisher_DbeqBasicDbeq:
		return Dataset_DbeqBasic
	case Publisher_ArcxPillarArcx:
		return Dataset_ArcxPillar
	case Publisher_IexgTopsIexg:
		return Dataset_IexgTops
	case Publisher_EqusPlusXchi, Publisher_EqusPlusXcis, Publisher_EqusPlusIexg, Publisher_EqusPlusEprl, Publisher_EqusPlusXnas, Publisher_EqusPlusXnys, Publisher_EqusPlusFinn, Publisher_EqusPlusFiny, Publisher_EqusPlusFinc, Publisher_EqusPlusEqus:
		return Dataset_EqusPlus
	case Publisher_XnysBboXnys:
		return Dataset_XnysBbo
	case Publisher_XnysTradesFiny, Publisher_XnysTradesXnys, Publisher_XnysTradesEqus:
		return Dataset_XnysTrades
	case Publisher_XnasQbboXnas:
		return Dataset_XnasQbbo
	case Publisher_XnasNlsFinn, Publisher_XnasNlsFinc, Publisher_XnasNlsXnas, Publisher_XnasNlsXbos, Publisher_XnasNlsXpsx:
		return Dataset_XnasNls
	case Publisher_IfeuImpactIfeu, Publisher_IfeuImpactXoff:
		return Dataset_IfeuImpact
	case Publisher_NdexImpactNdex, Publisher_NdexImpactXoff:
		return Dataset_NdexImpact
	case Publisher_EqusAllXchi, Publisher_EqusAllXcis, Publisher_EqusAllIexg, Publisher_EqusAllEprl, Publisher_EqusAllXnas, Publisher_EqusAllXnys, Publisher_EqusAllFinn, Publisher_EqusAllFiny, Publisher_EqusAllFinc, Publisher_EqusAllBats, Publisher_EqusAllBaty, Publisher_EqusAllEdga, Publisher_EqusAllEdgx, Publisher_EqusAllXbos, Publisher_EqusAllXpsx, Publisher_EqusAllMemx, Publisher_EqusAllXase, Publisher_EqusAllArcx, Publisher_EqusAllLtse, Publisher_EqusAllEqus:
		return Dataset_EqusAll
	case Publisher_XnasBasicXnas, Publisher_XnasBasicFinn, Publisher_XnasBasicFinc, Publisher_XnasBasicXbos, Publisher_XnasBasicXpsx, Publisher_XnasBasicEqus:
		return Dataset_XnasBasic
	case Publisher_EqusSummaryEqus:
		return Dataset_EqusSummary
	case Publisher_XcisTradesbboXcis:
		return Dataset_XcisTradesbbo
	case Publisher_XnysTradesbboXnys:
		return Dataset_XnysTradesbbo
	case Publisher_EqusMiniEqus:
		return Dataset_EqusMini
	case Publisher_IfusImpactIfus, Publisher_IfusImpactXoff:
		return Dataset_IfusImpact
	case Publisher_IfllImpactIfll, Publisher_IfllImpactXoff:
		return Dataset_IfllImpact
	case Publisher_XeurEobiXeur, Publisher_XeurEobiXoff:
		return Dataset_XeurEobi
	case Publisher_XeeeEobiXeee, Publisher_XeeeEobiXoff:
		return Dataset_XeeeEobi
	case Publisher_XcbfPitchXcbf, Publisher_XcbfPitchXoff:
		return Dataset_XcbfPitch
	case Publisher_OceaMemoirOcea:
		return Dataset_OceaMemoir
	case Publisher_MainCgifMain, Publisher_MainCgifCdef:
		return Dataset_MainCgif
	case Publisher_EqusSipXase, Publisher_EqusSipXbos, Publisher_EqusSipXcis, Publisher_EqusSipTxse, Publisher_EqusSip24Eq, Publisher_EqusSipEprl, Publisher_EqusSipXisx, Publisher_EqusSipEdga, Publisher_EqusSipEdgx, Publisher_EqusSipLtse, Publisher_EqusSipXchi, Publisher_EqusSipXnys, Publisher_EqusSipArcx, Publisher_EqusSipXnas, Publisher_EqusSipMemx, Publisher_EqusSipIexg, Publisher_EqusSipXpsx, Publisher_EqusSipBaty, Publisher_EqusSipBats, Publisher_EqusSipFinn, Publisher_EqusSipFiny, Publisher_EqusSipFinc, Publisher_EqusSipXadf, Publisher_EqusSipCta, Publisher_EqusSipUtp:
		return Dataset_EqusSip
	case Publisher_MsciCgifMsci:
		return Dataset_MsciCgif
	case Publisher_FtseCgifFtse:
		return Dataset_FtseCgif
	case Publisher_InavCgifInav:
		return Dataset_InavCgif
	case Publisher_MstarCgifMstar:
		return Dataset_MstarCgif
	case Publisher_CccyCgifCccy:
		return Dataset_CccyCgif
	case Publisher_CgiCgifCgi:
		return Dataset_CgiCgif
	case Publisher_XtksFlexXtks:
		return Dataset_XtksFlex
	case Publisher_XtktItchXtkt:
		return Dataset_XtktItch
	case Publisher_XoseItchXose:
		return Dataset_XoseItch
	case Publisher_DbixCoreMain, Publisher_DbixCoreMsci, Publisher_DbixCoreFtse, Publisher_DbixCoreInav, Publisher_DbixCoreMstar, Publisher_DbixCoreCccy, Publisher_DbixCoreCgi, Publisher_DbixCoreCdef:
		return Dataset_DbixCore
	default:
		return 0
	}
}

// PublisherFromDatasetVenue constructs a Publisher from its Dataset and Venue components.
// Returns an error if there's no Publisher with the corresponding Dataset and Venue combination.
func PublisherFromDatasetVenue(dataset Dataset, venue Venue) (Publisher, error) {
	switch dataset {
	case Dataset_GlbxMdp3:
		if venue == Venue_Glbx {
			return Publisher_GlbxMdp3Glbx, nil
		}
	case Dataset_XnasItch:
		if venue == Venue_Xnas {
			return Publisher_XnasItchXnas, nil
		}
	case Dataset_XbosItch:
		if venue == Venue_Xbos {
			return Publisher_XbosItchXbos, nil
		}
	case Dataset_XpsxItch:
		if venue == Venue_Xpsx {
			return Publisher_XpsxItchXpsx, nil
		}
	case Dataset_BatsPitch:
		if venue == Venue_Bats {
			return Publisher_BatsPitchBats, nil
		}
	case Dataset_BatyPitch:
		if venue == Venue_Baty {
			return Publisher_BatyPitchBaty, nil
		}
	case Dataset_EdgaPitch:
		if venue == Venue_Edga {
			return Publisher_EdgaPitchEdga, nil
		}
	case Dataset_EdgxPitch:
		if venue == Venue_Edgx {
			return Publisher_EdgxPitchEdgx, nil
		}
	case Dataset_XnysPillar:
		if venue == Venue_Xnys {
			return Publisher_XnysPillarXnys, nil
		}
	case Dataset_XcisPillar:
		if venue == Venue_Xcis {
			return Publisher_XcisPillarXcis, nil
		}
	case Dataset_XasePillar:
		if venue == Venue_Xase {
			return Publisher_XasePillarXase, nil
		}
	case Dataset_XchiPillar:
		if venue == Venue_Xchi {
			return Publisher_XchiPillarXchi, nil
		}
	case Dataset_XcisBbo:
		if venue == Venue_Xcis {
			return Publisher_XcisBboXcis, nil
		}
	case Dataset_XcisTrades:
		if venue == Venue_Xcis {
			return Publisher_XcisTradesXcis, nil
		}
	case Dataset_MemxMemoir:
		if venue == Venue_Memx {
			return Publisher_MemxMemoirMemx, nil
		}
	case Dataset_EprlDom:
		if venue == Venue_Eprl {
			return Publisher_EprlDomEprl, nil
		}
	case Dataset_OpraPillar:
		switch venue {
		case Venue_Amxo:
			return Publisher_OpraPillarAmxo, nil
		case Venue_Xbox:
			return Publisher_OpraPillarXbox, nil
		case Venue_Xcbo:
			return Publisher_OpraPillarXcbo, nil
		case Venue_Emld:
			return Publisher_OpraPillarEmld, nil
		case Venue_Edgo:
			return Publisher_OpraPillarEdgo, nil
		case Venue_Gmni:
			return Publisher_OpraPillarGmni, nil
		case Venue_Xisx:
			return Publisher_OpraPillarXisx, nil
		case Venue_Mcry:
			return Publisher_OpraPillarMcry, nil
		case Venue_Xmio:
			return Publisher_OpraPillarXmio, nil
		case Venue_Arco:
			return Publisher_OpraPillarArco, nil
		case Venue_Opra:
			return Publisher_OpraPillarOpra, nil
		case Venue_Mprl:
			return Publisher_OpraPillarMprl, nil
		case Venue_Xndq:
			return Publisher_OpraPillarXndq, nil
		case Venue_Xbxo:
			return Publisher_OpraPillarXbxo, nil
		case Venue_C2Ox:
			return Publisher_OpraPillarC2Ox, nil
		case Venue_Xphl:
			return Publisher_OpraPillarXphl, nil
		case Venue_Bato:
			return Publisher_OpraPillarBato, nil
		case Venue_Mxop:
			return Publisher_OpraPillarMxop, nil
		case Venue_Sphr:
			return Publisher_OpraPillarSphr, nil
		case Venue_Mxto:
			return Publisher_OpraPillarMxto, nil
		case Venue_Iexo:
			return Publisher_OpraPillarIexo, nil
		}
	case Dataset_DbeqBasic:
		switch venue {
		case Venue_Xchi:
			return Publisher_DbeqBasicXchi, nil
		case Venue_Xcis:
			return Publisher_DbeqBasicXcis, nil
		case Venue_Iexg:
			return Publisher_DbeqBasicIexg, nil
		case Venue_Eprl:
			return Publisher_DbeqBasicEprl, nil
		case Venue_Dbeq:
			return Publisher_DbeqBasicDbeq, nil
		}
	case Dataset_ArcxPillar:
		if venue == Venue_Arcx {
			return Publisher_ArcxPillarArcx, nil
		}
	case Dataset_IexgTops:
		if venue == Venue_Iexg {
			return Publisher_IexgTopsIexg, nil
		}
	case Dataset_EqusPlus:
		switch venue {
		case Venue_Xchi:
			return Publisher_EqusPlusXchi, nil
		case Venue_Xcis:
			return Publisher_EqusPlusXcis, nil
		case Venue_Iexg:
			return Publisher_EqusPlusIexg, nil
		case Venue_Eprl:
			return Publisher_EqusPlusEprl, nil
		case Venue_Xnas:
			return Publisher_EqusPlusXnas, nil
		case Venue_Xnys:
			return Publisher_EqusPlusXnys, nil
		case Venue_Finn:
			return Publisher_EqusPlusFinn, nil
		case Venue_Finy:
			return Publisher_EqusPlusFiny, nil
		case Venue_Finc:
			return Publisher_EqusPlusFinc, nil
		case Venue_Equs:
			return Publisher_EqusPlusEqus, nil
		}
	case Dataset_XnysBbo:
		if venue == Venue_Xnys {
			return Publisher_XnysBboXnys, nil
		}
	case Dataset_XnysTrades:
		switch venue {
		case Venue_Finy:
			return Publisher_XnysTradesFiny, nil
		case Venue_Xnys:
			return Publisher_XnysTradesXnys, nil
		case Venue_Equs:
			return Publisher_XnysTradesEqus, nil
		}
	case Dataset_XnasQbbo:
		if venue == Venue_Xnas {
			return Publisher_XnasQbboXnas, nil
		}
	case Dataset_XnasNls:
		switch venue {
		case Venue_Finn:
			return Publisher_XnasNlsFinn, nil
		case Venue_Finc:
			return Publisher_XnasNlsFinc, nil
		case Venue_Xnas:
			return Publisher_XnasNlsXnas, nil
		case Venue_Xbos:
			return Publisher_XnasNlsXbos, nil
		case Venue_Xpsx:
			return Publisher_XnasNlsXpsx, nil
		}
	case Dataset_IfeuImpact:
		switch venue {
		case Venue_Ifeu:
			return Publisher_IfeuImpactIfeu, nil
		case Venue_Xoff:
			return Publisher_IfeuImpactXoff, nil
		}
	case Dataset_NdexImpact:
		switch venue {
		case Venue_Ndex:
			return Publisher_NdexImpactNdex, nil
		case Venue_Xoff:
			return Publisher_NdexImpactXoff, nil
		}
	case Dataset_EqusAll:
		switch venue {
		case Venue_Xchi:
			return Publisher_EqusAllXchi, nil
		case Venue_Xcis:
			return Publisher_EqusAllXcis, nil
		case Venue_Iexg:
			return Publisher_EqusAllIexg, nil
		case Venue_Eprl:
			return Publisher_EqusAllEprl, nil
		case Venue_Xnas:
			return Publisher_EqusAllXnas, nil
		case Venue_Xnys:
			return Publisher_EqusAllXnys, nil
		case Venue_Finn:
			return Publisher_EqusAllFinn, nil
		case Venue_Finy:
			return Publisher_EqusAllFiny, nil
		case Venue_Finc:
			return Publisher_EqusAllFinc, nil
		case Venue_Bats:
			return Publisher_EqusAllBats, nil
		case Venue_Baty:
			return Publisher_EqusAllBaty, nil
		case Venue_Edga:
			return Publisher_EqusAllEdga, nil
		case Venue_Edgx:
			return Publisher_EqusAllEdgx, nil
		case Venue_Xbos:
			return Publisher_EqusAllXbos, nil
		case Venue_Xpsx:
			return Publisher_EqusAllXpsx, nil
		case Venue_Memx:
			return Publisher_EqusAllMemx, nil
		case Venue_Xase:
			return Publisher_EqusAllXase, nil
		case Venue_Arcx:
			return Publisher_EqusAllArcx, nil
		case Venue_Ltse:
			return Publisher_EqusAllLtse, nil
		case Venue_Equs:
			return Publisher_EqusAllEqus, nil
		}
	case Dataset_XnasBasic:
		switch venue {
		case Venue_Xnas:
			return Publisher_XnasBasicXnas, nil
		case Venue_Finn:
			return Publisher_XnasBasicFinn, nil
		case Venue_Finc:
			return Publisher_XnasBasicFinc, nil
		case Venue_Xbos:
			return Publisher_XnasBasicXbos, nil
		case Venue_Xpsx:
			return Publisher_XnasBasicXpsx, nil
		case Venue_Equs:
			return Publisher_XnasBasicEqus, nil
		}
	case Dataset_EqusSummary:
		if venue == Venue_Equs {
			return Publisher_EqusSummaryEqus, nil
		}
	case Dataset_XcisTradesbbo:
		if venue == Venue_Xcis {
			return Publisher_XcisTradesbboXcis, nil
		}
	case Dataset_XnysTradesbbo:
		if venue == Venue_Xnys {
			return Publisher_XnysTradesbboXnys, nil
		}
	case Dataset_EqusMini:
		if venue == Venue_Equs {
			return Publisher_EqusMiniEqus, nil
		}
	case Dataset_IfusImpact:
		switch venue {
		case Venue_Ifus:
			return Publisher_IfusImpactIfus, nil
		case Venue_Xoff:
			return Publisher_IfusImpactXoff, nil
		}
	case Dataset_IfllImpact:
		switch venue {
		case Venue_Ifll:
			return Publisher_IfllImpactIfll, nil
		case Venue_Xoff:
			return Publisher_IfllImpactXoff, nil
		}
	case Dataset_XeurEobi:
		switch venue {
		case Venue_Xeur:
			return Publisher_XeurEobiXeur, nil
		case Venue_Xoff:
			return Publisher_XeurEobiXoff, nil
		}
	case Dataset_XeeeEobi:
		switch venue {
		case Venue_Xeee:
			return Publisher_XeeeEobiXeee, nil
		case Venue_Xoff:
			return Publisher_XeeeEobiXoff, nil
		}
	case Dataset_XcbfPitch:
		switch venue {
		case Venue_Xcbf:
			return Publisher_XcbfPitchXcbf, nil
		case Venue_Xoff:
			return Publisher_XcbfPitchXoff, nil
		}
	case Dataset_OceaMemoir:
		if venue == Venue_Ocea {
			return Publisher_OceaMemoirOcea, nil
		}
	case Dataset_MainCgif:
		switch venue {
		case Venue_Main:
			return Publisher_MainCgifMain, nil
		case Venue_Cdef:
			return Publisher_MainCgifCdef, nil
		}
	case Dataset_EqusSip:
		switch venue {
		case Venue_Xase:
			return Publisher_EqusSipXase, nil
		case Venue_Xbos:
			return Publisher_EqusSipXbos, nil
		case Venue_Xcis:
			return Publisher_EqusSipXcis, nil
		case Venue_Txse:
			return Publisher_EqusSipTxse, nil
		case Venue__24Eq:
			return Publisher_EqusSip24Eq, nil
		case Venue_Eprl:
			return Publisher_EqusSipEprl, nil
		case Venue_Xisx:
			return Publisher_EqusSipXisx, nil
		case Venue_Edga:
			return Publisher_EqusSipEdga, nil
		case Venue_Edgx:
			return Publisher_EqusSipEdgx, nil
		case Venue_Ltse:
			return Publisher_EqusSipLtse, nil
		case Venue_Xchi:
			return Publisher_EqusSipXchi, nil
		case Venue_Xnys:
			return Publisher_EqusSipXnys, nil
		case Venue_Arcx:
			return Publisher_EqusSipArcx, nil
		case Venue_Xnas:
			return Publisher_EqusSipXnas, nil
		case Venue_Memx:
			return Publisher_EqusSipMemx, nil
		case Venue_Iexg:
			return Publisher_EqusSipIexg, nil
		case Venue_Xpsx:
			return Publisher_EqusSipXpsx, nil
		case Venue_Baty:
			return Publisher_EqusSipBaty, nil
		case Venue_Bats:
			return Publisher_EqusSipBats, nil
		case Venue_Finn:
			return Publisher_EqusSipFinn, nil
		case Venue_Finy:
			return Publisher_EqusSipFiny, nil
		case Venue_Finc:
			return Publisher_EqusSipFinc, nil
		case Venue_Xadf:
			return Publisher_EqusSipXadf, nil
		case Venue_Cta:
			return Publisher_EqusSipCta, nil
		case Venue_Utp:
			return Publisher_EqusSipUtp, nil
		}
	case Dataset_MsciCgif:
		if venue == Venue_Msci {
			return Publisher_MsciCgifMsci, nil
		}
	case Dataset_FtseCgif:
		if venue == Venue_Ftse {
			return Publisher_FtseCgifFtse, nil
		}
	case Dataset_InavCgif:
		if venue == Venue_Inav {
			return Publisher_InavCgifInav, nil
		}
	case Dataset_MstarCgif:
		if venue == Venue_Mstar {
			return Publisher_MstarCgifMstar, nil
		}
	case Dataset_CccyCgif:
		if venue == Venue_Cccy {
			return Publisher_CccyCgifCccy, nil
		}
	case Dataset_CgiCgif:
		if venue == Venue_Cgi {
			return Publisher_CgiCgifCgi, nil
		}
	case Dataset_XtksFlex:
		if venue == Venue_Xtks {
			return Publisher_XtksFlexXtks, nil
		}
	case Dataset_XtktItch:
		if venue == Venue_Xtkt {
			return Publisher_XtktItchXtkt, nil
		}
	case Dataset_XoseItch:
		if venue == Venue_Xose {
			return Publisher_XoseItchXose, nil
		}
	case Dataset_DbixCore:
		switch venue {
		case Venue_Main:
			return Publisher_DbixCoreMain, nil
		case Venue_Msci:
			return Publisher_DbixCoreMsci, nil
		case Venue_Ftse:
			return Publisher_DbixCoreFtse, nil
		case Venue_Inav:
			return Publisher_DbixCoreInav, nil
		case Venue_Mstar:
			return Publisher_DbixCoreMstar, nil
		case Venue_Cccy:
			return Publisher_DbixCoreCccy, nil
		case Venue_Cgi:
			return Publisher_DbixCoreCgi, nil
		case Venue_Cdef:
			return Publisher_DbixCoreCdef, nil
		}
	}
	return 0, fmt.Errorf("no publisher for dataset %s and venue %s", dataset.String(), venue.String())
}

// PublisherFromString converts a string to a Publisher.
// Returns an error if the string is unknown.
func PublisherFromString(str string) (Publisher, error) {
	str = strings.ToUpper(str)
	switch str {
	case "GLBX.MDP3.GLBX":
		return Publisher_GlbxMdp3Glbx, nil
	case "XNAS.ITCH.XNAS":
		return Publisher_XnasItchXnas, nil
	case "XBOS.ITCH.XBOS":
		return Publisher_XbosItchXbos, nil
	case "XPSX.ITCH.XPSX":
		return Publisher_XpsxItchXpsx, nil
	case "BATS.PITCH.BATS":
		return Publisher_BatsPitchBats, nil
	case "BATY.PITCH.BATY":
		return Publisher_BatyPitchBaty, nil
	case "EDGA.PITCH.EDGA":
		return Publisher_EdgaPitchEdga, nil
	case "EDGX.PITCH.EDGX":
		return Publisher_EdgxPitchEdgx, nil
	case "XNYS.PILLAR.XNYS":
		return Publisher_XnysPillarXnys, nil
	case "XCIS.PILLAR.XCIS":
		return Publisher_XcisPillarXcis, nil
	case "XASE.PILLAR.XASE":
		return Publisher_XasePillarXase, nil
	case "XCHI.PILLAR.XCHI":
		return Publisher_XchiPillarXchi, nil
	case "XCIS.BBO.XCIS":
		return Publisher_XcisBboXcis, nil
	case "XCIS.TRADES.XCIS":
		return Publisher_XcisTradesXcis, nil
	case "MEMX.MEMOIR.MEMX":
		return Publisher_MemxMemoirMemx, nil
	case "EPRL.DOM.EPRL":
		return Publisher_EprlDomEprl, nil
	case "XNAS.NLS.FINN":
		return Publisher_XnasNlsFinn, nil
	case "XNAS.NLS.FINC":
		return Publisher_XnasNlsFinc, nil
	case "XNYS.TRADES.FINY":
		return Publisher_XnysTradesFiny, nil
	case "OPRA.PILLAR.AMXO":
		return Publisher_OpraPillarAmxo, nil
	case "OPRA.PILLAR.XBOX":
		return Publisher_OpraPillarXbox, nil
	case "OPRA.PILLAR.XCBO":
		return Publisher_OpraPillarXcbo, nil
	case "OPRA.PILLAR.EMLD":
		return Publisher_OpraPillarEmld, nil
	case "OPRA.PILLAR.EDGO":
		return Publisher_OpraPillarEdgo, nil
	case "OPRA.PILLAR.GMNI":
		return Publisher_OpraPillarGmni, nil
	case "OPRA.PILLAR.XISX":
		return Publisher_OpraPillarXisx, nil
	case "OPRA.PILLAR.MCRY":
		return Publisher_OpraPillarMcry, nil
	case "OPRA.PILLAR.XMIO":
		return Publisher_OpraPillarXmio, nil
	case "OPRA.PILLAR.ARCO":
		return Publisher_OpraPillarArco, nil
	case "OPRA.PILLAR.OPRA":
		return Publisher_OpraPillarOpra, nil
	case "OPRA.PILLAR.MPRL":
		return Publisher_OpraPillarMprl, nil
	case "OPRA.PILLAR.XNDQ":
		return Publisher_OpraPillarXndq, nil
	case "OPRA.PILLAR.XBXO":
		return Publisher_OpraPillarXbxo, nil
	case "OPRA.PILLAR.C2OX":
		return Publisher_OpraPillarC2Ox, nil
	case "OPRA.PILLAR.XPHL":
		return Publisher_OpraPillarXphl, nil
	case "OPRA.PILLAR.BATO":
		return Publisher_OpraPillarBato, nil
	case "OPRA.PILLAR.MXOP":
		return Publisher_OpraPillarMxop, nil
	case "IEXG.TOPS.IEXG":
		return Publisher_IexgTopsIexg, nil
	case "DBEQ.BASIC.XCHI":
		return Publisher_DbeqBasicXchi, nil
	case "DBEQ.BASIC.XCIS":
		return Publisher_DbeqBasicXcis, nil
	case "DBEQ.BASIC.IEXG":
		return Publisher_DbeqBasicIexg, nil
	case "DBEQ.BASIC.EPRL":
		return Publisher_DbeqBasicEprl, nil
	case "ARCX.PILLAR.ARCX":
		return Publisher_ArcxPillarArcx, nil
	case "XNYS.BBO.XNYS":
		return Publisher_XnysBboXnys, nil
	case "XNYS.TRADES.XNYS":
		return Publisher_XnysTradesXnys, nil
	case "XNAS.QBBO.XNAS":
		return Publisher_XnasQbboXnas, nil
	case "XNAS.NLS.XNAS":
		return Publisher_XnasNlsXnas, nil
	case "EQUS.PLUS.XCHI":
		return Publisher_EqusPlusXchi, nil
	case "EQUS.PLUS.XCIS":
		return Publisher_EqusPlusXcis, nil
	case "EQUS.PLUS.IEXG":
		return Publisher_EqusPlusIexg, nil
	case "EQUS.PLUS.EPRL":
		return Publisher_EqusPlusEprl, nil
	case "EQUS.PLUS.XNAS":
		return Publisher_EqusPlusXnas, nil
	case "EQUS.PLUS.XNYS":
		return Publisher_EqusPlusXnys, nil
	case "EQUS.PLUS.FINN":
		return Publisher_EqusPlusFinn, nil
	case "EQUS.PLUS.FINY":
		return Publisher_EqusPlusFiny, nil
	case "EQUS.PLUS.FINC":
		return Publisher_EqusPlusFinc, nil
	case "IFEU.IMPACT.IFEU":
		return Publisher_IfeuImpactIfeu, nil
	case "NDEX.IMPACT.NDEX":
		return Publisher_NdexImpactNdex, nil
	case "DBEQ.BASIC.DBEQ":
		return Publisher_DbeqBasicDbeq, nil
	case "EQUS.PLUS.EQUS":
		return Publisher_EqusPlusEqus, nil
	case "OPRA.PILLAR.SPHR":
		return Publisher_OpraPillarSphr, nil
	case "EQUS.ALL.XCHI":
		return Publisher_EqusAllXchi, nil
	case "EQUS.ALL.XCIS":
		return Publisher_EqusAllXcis, nil
	case "EQUS.ALL.IEXG":
		return Publisher_EqusAllIexg, nil
	case "EQUS.ALL.EPRL":
		return Publisher_EqusAllEprl, nil
	case "EQUS.ALL.XNAS":
		return Publisher_EqusAllXnas, nil
	case "EQUS.ALL.XNYS":
		return Publisher_EqusAllXnys, nil
	case "EQUS.ALL.FINN":
		return Publisher_EqusAllFinn, nil
	case "EQUS.ALL.FINY":
		return Publisher_EqusAllFiny, nil
	case "EQUS.ALL.FINC":
		return Publisher_EqusAllFinc, nil
	case "EQUS.ALL.BATS":
		return Publisher_EqusAllBats, nil
	case "EQUS.ALL.BATY":
		return Publisher_EqusAllBaty, nil
	case "EQUS.ALL.EDGA":
		return Publisher_EqusAllEdga, nil
	case "EQUS.ALL.EDGX":
		return Publisher_EqusAllEdgx, nil
	case "EQUS.ALL.XBOS":
		return Publisher_EqusAllXbos, nil
	case "EQUS.ALL.XPSX":
		return Publisher_EqusAllXpsx, nil
	case "EQUS.ALL.MEMX":
		return Publisher_EqusAllMemx, nil
	case "EQUS.ALL.XASE":
		return Publisher_EqusAllXase, nil
	case "EQUS.ALL.ARCX":
		return Publisher_EqusAllArcx, nil
	case "EQUS.ALL.LTSE":
		return Publisher_EqusAllLtse, nil
	case "XNAS.BASIC.XNAS":
		return Publisher_XnasBasicXnas, nil
	case "XNAS.BASIC.FINN":
		return Publisher_XnasBasicFinn, nil
	case "XNAS.BASIC.FINC":
		return Publisher_XnasBasicFinc, nil
	case "IFEU.IMPACT.XOFF":
		return Publisher_IfeuImpactXoff, nil
	case "NDEX.IMPACT.XOFF":
		return Publisher_NdexImpactXoff, nil
	case "XNAS.NLS.XBOS":
		return Publisher_XnasNlsXbos, nil
	case "XNAS.NLS.XPSX":
		return Publisher_XnasNlsXpsx, nil
	case "XNAS.BASIC.XBOS":
		return Publisher_XnasBasicXbos, nil
	case "XNAS.BASIC.XPSX":
		return Publisher_XnasBasicXpsx, nil
	case "EQUS.SUMMARY.EQUS":
		return Publisher_EqusSummaryEqus, nil
	case "XCIS.TRADESBBO.XCIS":
		return Publisher_XcisTradesbboXcis, nil
	case "XNYS.TRADESBBO.XNYS":
		return Publisher_XnysTradesbboXnys, nil
	case "XNAS.BASIC.EQUS":
		return Publisher_XnasBasicEqus, nil
	case "EQUS.ALL.EQUS":
		return Publisher_EqusAllEqus, nil
	case "EQUS.MINI.EQUS":
		return Publisher_EqusMiniEqus, nil
	case "XNYS.TRADES.EQUS":
		return Publisher_XnysTradesEqus, nil
	case "IFUS.IMPACT.IFUS":
		return Publisher_IfusImpactIfus, nil
	case "IFUS.IMPACT.XOFF":
		return Publisher_IfusImpactXoff, nil
	case "IFLL.IMPACT.IFLL":
		return Publisher_IfllImpactIfll, nil
	case "IFLL.IMPACT.XOFF":
		return Publisher_IfllImpactXoff, nil
	case "XEUR.EOBI.XEUR":
		return Publisher_XeurEobiXeur, nil
	case "XEEE.EOBI.XEEE":
		return Publisher_XeeeEobiXeee, nil
	case "XEUR.EOBI.XOFF":
		return Publisher_XeurEobiXoff, nil
	case "XEEE.EOBI.XOFF":
		return Publisher_XeeeEobiXoff, nil
	case "XCBF.PITCH.XCBF":
		return Publisher_XcbfPitchXcbf, nil
	case "XCBF.PITCH.XOFF":
		return Publisher_XcbfPitchXoff, nil
	case "OCEA.MEMOIR.OCEA":
		return Publisher_OceaMemoirOcea, nil
	case "OPRA.PILLAR.MXTO":
		return Publisher_OpraPillarMxto, nil
	case "OPRA.PILLAR.IEXO":
		return Publisher_OpraPillarIexo, nil
	case "MAIN.CGIF.MAIN":
		return Publisher_MainCgifMain, nil
	case "EQUS.SIP.XASE":
		return Publisher_EqusSipXase, nil
	case "EQUS.SIP.XBOS":
		return Publisher_EqusSipXbos, nil
	case "EQUS.SIP.XCIS":
		return Publisher_EqusSipXcis, nil
	case "EQUS.SIP.TXSE":
		return Publisher_EqusSipTxse, nil
	case "EQUS.SIP.24EQ":
		return Publisher_EqusSip24Eq, nil
	case "EQUS.SIP.EPRL":
		return Publisher_EqusSipEprl, nil
	case "EQUS.SIP.XISX":
		return Publisher_EqusSipXisx, nil
	case "EQUS.SIP.EDGA":
		return Publisher_EqusSipEdga, nil
	case "EQUS.SIP.EDGX":
		return Publisher_EqusSipEdgx, nil
	case "EQUS.SIP.LTSE":
		return Publisher_EqusSipLtse, nil
	case "EQUS.SIP.XCHI":
		return Publisher_EqusSipXchi, nil
	case "EQUS.SIP.XNYS":
		return Publisher_EqusSipXnys, nil
	case "EQUS.SIP.ARCX":
		return Publisher_EqusSipArcx, nil
	case "EQUS.SIP.XNAS":
		return Publisher_EqusSipXnas, nil
	case "EQUS.SIP.MEMX":
		return Publisher_EqusSipMemx, nil
	case "EQUS.SIP.IEXG":
		return Publisher_EqusSipIexg, nil
	case "EQUS.SIP.XPSX":
		return Publisher_EqusSipXpsx, nil
	case "EQUS.SIP.BATY":
		return Publisher_EqusSipBaty, nil
	case "EQUS.SIP.BATS":
		return Publisher_EqusSipBats, nil
	case "EQUS.SIP.FINN":
		return Publisher_EqusSipFinn, nil
	case "EQUS.SIP.FINY":
		return Publisher_EqusSipFiny, nil
	case "EQUS.SIP.FINC":
		return Publisher_EqusSipFinc, nil
	case "EQUS.SIP.XADF":
		return Publisher_EqusSipXadf, nil
	case "EQUS.SIP.CTA":
		return Publisher_EqusSipCta, nil
	case "EQUS.SIP.UTP":
		return Publisher_EqusSipUtp, nil
	case "MSCI.CGIF.MSCI":
		return Publisher_MsciCgifMsci, nil
	case "FTSE.CGIF.FTSE":
		return Publisher_FtseCgifFtse, nil
	case "INAV.CGIF.INAV":
		return Publisher_InavCgifInav, nil
	case "MSTAR.CGIF.MSTAR":
		return Publisher_MstarCgifMstar, nil
	case "CCCY.CGIF.CCCY":
		return Publisher_CccyCgifCccy, nil
	case "CGI.CGIF.CGI":
		return Publisher_CgiCgifCgi, nil
	case "MAIN.CGIF.CDEF":
		return Publisher_MainCgifCdef, nil
	case "XTKS.FLEX.XTKS":
		return Publisher_XtksFlexXtks, nil
	case "XTKT.ITCH.XTKT":
		return Publisher_XtktItchXtkt, nil
	case "XOSE.ITCH.XOSE":
		return Publisher_XoseItchXose, nil
	case "DBIX.CORE.MAIN":
		return Publisher_DbixCoreMain, nil
	case "DBIX.CORE.MSCI":
		return Publisher_DbixCoreMsci, nil
	case "DBIX.CORE.FTSE":
		return Publisher_DbixCoreFtse, nil
	case "DBIX.CORE.INAV":
		return Publisher_DbixCoreInav, nil
	case "DBIX.CORE.MSTAR":
		return Publisher_DbixCoreMstar, nil
	case "DBIX.CORE.CCCY":
		return Publisher_DbixCoreCccy, nil
	case "DBIX.CORE.CGI":
		return Publisher_DbixCoreCgi, nil
	case "DBIX.CORE.CDEF":
		return Publisher_DbixCoreCdef, nil
	default:
		return 0, fmt.Errorf("unknown publisher: '%s'", str)
	}
}

func (p Publisher) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *Publisher) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	pub, err := PublisherFromString(str)
	if err != nil {
		return err
	}
	*p = pub
	return nil
}

// Type implements pflag.Value.Type. Returns "dbn.Publisher".
func (*Publisher) Type() string {
	return "dbn.Publisher"
}

// Set implements the flag.Value interface.
func (p *Publisher) Set(value string) error {
	pub, err := PublisherFromString(value)
	if err == nil {
		*p = pub
	}
	return err
}
