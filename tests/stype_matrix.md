# Databento Symbology Support Matrix

Generated: 2026-02-11

Probed using `symbology.resolve` with dummy symbols. A checkmark means the API
accepted the stype for that dataset (HTTP 200). An `x` means the API rejected
the combination as unsupported (HTTP 422 `symbology_invalid_request`).

## stype_in support

Each cell shows whether `stype_in=<column>` is accepted when `stype_out=instrument_id`.

| Dataset | raw_symbol | instrument_id | smart | continuous | parent | nasdaq_symbol | cms_symbol | isin | us_code | bbg_comp_id | bbg_comp_ticker | figi | figi_ticker |
| --- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| ARCX.PILLAR | Y | Y | - | - | - | - | - | - | - | - | - | - | - |
| BATS.PITCH | Y | Y | - | - | - | - | - | - | - | - | - | - | - |
| BATY.PITCH | Y | Y | - | - | - | - | - | - | - | - | - | - | - |
| DBEQ.BASIC | Y | Y | - | - | - | - | - | - | - | - | - | - | - |
| EDGA.PITCH | Y | Y | - | - | - | - | - | - | - | - | - | - | - |
| EDGX.PITCH | Y | Y | - | - | - | - | - | - | - | - | - | - | - |
| EPRL.DOM | Y | Y | - | - | - | - | - | - | - | - | - | - | - |
| EQUS.MINI | Y | Y | - | - | - | - | - | - | - | - | - | - | - |
| EQUS.SUMMARY | Y | Y | - | - | - | - | - | - | - | - | - | - | - |
| GLBX.MDP3 | Y | Y | - | Y | - | - | - | - | - | - | - | - | - |
| IEXG.TOPS | Y | Y | - | - | - | - | - | - | - | - | - | - | - |
| IFEU.IMPACT | Y | Y | - | Y | - | - | - | - | - | - | - | - | - |
| IFLL.IMPACT | Y | Y | - | Y | - | - | - | - | - | - | - | - | - |
| IFUS.IMPACT | Y | Y | - | Y | - | - | - | - | - | - | - | - | - |
| MEMX.MEMOIR | Y | Y | - | - | - | - | - | - | - | - | - | - | - |
| NDEX.IMPACT | Y | Y | - | Y | - | - | - | - | - | - | - | - | - |
| OPRA.PILLAR | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| XASE.PILLAR | Y | Y | - | - | - | - | - | - | - | - | - | - | - |
| XBOS.ITCH | Y | Y | - | - | - | - | - | - | - | - | - | - | - |
| XCHI.PILLAR | Y | Y | - | - | - | - | - | - | - | - | - | - | - |
| XCIS.TRADESBBO | Y | Y | - | - | - | - | - | - | - | - | - | - | - |
| XEEE.EOBI | - | - | - | - | - | - | - | - | - | - | - | - | - |
| XEUR.EOBI | - | - | - | - | - | - | - | - | - | - | - | - | - |
| XNAS.BASIC | Y | Y | - | - | - | - | - | - | - | - | - | - | - |
| XNAS.ITCH | Y | Y | - | - | - | - | - | - | - | - | - | - | - |
| XNYS.PILLAR | Y | Y | - | - | - | - | - | - | - | - | - | - | - |
| XPSX.ITCH | Y | Y | - | - | - | - | - | - | - | - | - | - | - |

## stype_out support

Each cell shows whether `stype_out=<column>` is accepted when `stype_in=raw_symbol`.

| Dataset | raw_symbol | instrument_id | smart | continuous | parent | nasdaq_symbol | cms_symbol | isin | us_code | bbg_comp_id | bbg_comp_ticker | figi | figi_ticker |
| --- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| ARCX.PILLAR | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| BATS.PITCH | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| BATY.PITCH | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| DBEQ.BASIC | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| EDGA.PITCH | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| EDGX.PITCH | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| EPRL.DOM | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| EQUS.MINI | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| EQUS.SUMMARY | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| GLBX.MDP3 | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| IEXG.TOPS | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| IFEU.IMPACT | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| IFLL.IMPACT | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| IFUS.IMPACT | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| MEMX.MEMOIR | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| NDEX.IMPACT | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| OPRA.PILLAR | - | - | - | - | - | - | - | - | - | - | - | - | - |
| XASE.PILLAR | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| XBOS.ITCH | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| XCHI.PILLAR | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| XCIS.TRADESBBO | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| XEEE.EOBI | - | - | - | - | - | - | - | - | - | - | - | - | - |
| XEUR.EOBI | - | - | - | - | - | - | - | - | - | - | - | - | - |
| XNAS.BASIC | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| XNAS.ITCH | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| XNYS.PILLAR | - | Y | - | - | - | - | - | - | - | - | - | - | - |
| XPSX.ITCH | - | Y | - | - | - | - | - | - | - | - | - | - | - |
