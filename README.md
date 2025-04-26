# StateStinger: 
**StateStinger** is a fuzzing engine focused on **state machine security** in Go-powered blockchains like Cosmos SDK. It injects randomized or malformed inputs into key transaction and state transition logic to uncover **critical bugs**, **invalid state transitions**, and **consensus-breaking behaviors**.


Ci:Passin {Icon}

## Features 

### ⚙️ What It Targets

Currently, StateStinger focuses on:

- ✅ **Cosmos SDK apps**
  - `DeliverTx`, `BeginBlock`, `EndBlock`
  - Built-in modules like `x/bank`, `x/staking`, `x/auth`

Planned support:
- ☑️ Ethermint
- ☑️ Tendermint Core fuzzing
- ☑️ Other Go-based blockchains (e.g., Celestia Core, Hyperledger Fabric)


## Installation

## Usage 

## Limitations and known issues

## License
MIT

## 🧠 About the Author
- Zakaria — Blockchain Engineer @ Go Sec Labs
- Focus: distributed systems security, protocol security, and tooling for next-gen blockchains.
