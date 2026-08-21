# UTLS

is a proxy tool wrapping around the "utls" go library.

Every connection to `$INPUT` Unix stream socket gets a respective connection to `$OUTPUT` Unix stream socket. The `$INPUT` connection has plaintext, yet `$OUTPUT` receives utls-encrypted connections.

The first argument to this program is the name of the fingerprint to use. List of fingerprints mirrors that of utls, taken from the variables from the "utls/u_common.go" file, just without the "Hello" prefix: for example, the fingerprint for the latest Firefox version is called "Firefox_Auto" in this tool, while in that file it is called "HelloFirefox_Auto"

## Names of supported fingerprints

~~(I hope I haven't forgotten to update it)~~

* Golang
* Randomized
* RandomizedALPN
* RandomizedNoALPN
* Firefox_Auto
* Firefox_55
* Firefox_56
* Firefox_63
* Firefox_65
* Firefox_99
* Firefox_102
* Firefox_105
* Firefox_120
* Firefox_148
* Chrome_Auto
* Chrome_58
* Chrome_62
* Chrome_70
* Chrome_72
* Chrome_83
* Chrome_87
* Chrome_96
* Chrome_100
* Chrome_102
* Chrome_106_Shuffle
* Chrome_100_PSK
* Chrome_112_PSK_Shuf
* Chrome_114_Padding_PSK_Shuf
* Chrome_115_PQ
* Chrome_115_PQ_PSK
* Chrome_120
* Chrome_120_PQ
* Chrome_131
* Chrome_133
* IOS_Auto
* IOS_11_1
* IOS_12_1
* IOS_13
* IOS_14
* Android_11_OkHttp
* Edge_Auto
* Edge_85
* Edge_106
* Safari_Auto
* Safari_16_0
* Safari_26_3
* 360_Auto
* 360_7_5
* 360_11_0
* QQ_Auto
* QQ_11_1

## Example use

```
env INPUT=./input_socket_path OUTPUT=./output_socket_path utls Firefox_Auto
```
