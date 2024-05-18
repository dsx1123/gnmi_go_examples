# Golang gNMI examples

a golang application to demostrate how to use gNMI to automate Cisco NX-OS

### Examples of the intput config:

```yaml
---
address: 172.31.186.156:50050                   # in the format of <host>:<port>
username: shdu
password: password
tls_ca: ca-chain.cert.pem
tls_cert: user_shdu_chain.cert.pem              # Must be a certificate chain
tls_key: user_shdu.key.unenc.pem                # Must be unencrypted private key
encoding: PROTO
get: /interfaces/interface[name=eth1/1]/state
set:
  path: /System
  file: ./config/int_desc_update.json
replace:
  path: /System/bgp-items
  file: ./config/bgp_items.json
subscriptions:
  - path: /interfaces/interface/state/oper-status
    mode: ON_CHANGE                             # ON_CHANGE, SAMPLE or TARGET_DEFINED
```

### How to run the examples:

```shell
A demo application of gNMI:
        Demonstrate gNMI  CAPABILITES/GET/SET/SUBSCRIBE.

Usage:
  gnmi_go [command]

Available Commands:
  cap         Get the gNMI capabilites from the target
  completion  Generate the autocompletion script for the specified shell
  delete      Delete a conatiner or leaf on the target
  eda         run EDA demo on target
  get         Get the configuration or operational state from the target
  help        Help about any command
  merge       Merge the candidate configuration wtih the configuration on target
  replace     Replace the configurations on the target
  subscribe   run gnmi subscribe on target

Flags:
      --config string   target config file (default "./config.yaml")
  -h, --help            help for gnmi_go

Use "gnmi_go [command] --help" for more information about a command.
```
