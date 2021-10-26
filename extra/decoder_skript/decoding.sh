#!/bin/bash
#
# von base64 nach HEX zeilenweise
echo BD-Sensor decoding
echo komplettes Telegramm
echo ohne Fehlererkennung im Payload
echo 
cat ./input.base64 | base64 -d  | xxd   -ps  -c1 > ./hex 

# zum prüfen ob Payload vollständig
len=$(cat hex | wc -l)
echo Anzahl Bytes im Payload : $len

#HEX \n ersetzt durch ' '
tr  '\n' ' ' < ./hex > ./hex2



#Adresse von dem Sensor in HEX
echo Adresse Sensor .........: 0x$(cut -d' ' -f48 hex2)

# Seriennummer zum prüfen ob es der richtige Sensor ist
SN1=$(cut -d' ' -f55 hex2)
SN2=$(cut -d' ' -f54 hex2)
SN3=$(cut -d' ' -f61 hex2)
SN4=$(cut -d' ' -f60 hex2)
echo Seriennummer ...........: $(( 16#$SN1$SN2$SN3$SN4 ))


# Actual pressure AP
AP1=$(cut -d' ' -f67 hex2)  # Achtung hier ist die Reihenfolge anders 
AP2=$(cut -d' ' -f66 hex2)  # noch anzupassen
AP3=$(cut -d' ' -f73 hex2)
AP4=$(cut -d' ' -f72 hex2)
#echo $AP4 $AP3 $AP2 $AP1
AP=$(echo -ne "\x"$AP4"\x"$AP3"\x"$AP2"\x"$AP1 | hexdump -e '1/4 "%f" "\n"')
echo Actual pressure .........: $AP