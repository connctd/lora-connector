#!/bin/sh

# von base64 nach HEX zeilenweise
echo BD-Sensor decoding
echo ohne Fehlererkennung im Payload
echo 
cat ./input.base64 | base64 -d  | xxd   -ps  -c1 > ./hex 

# zum prüfen ob Payload vollständig
len=$(cat hex | wc -l)
echo Anzahl Bytes im Payload : $len

#HEX \n ersetzt durch ' '
tr  '\n' ' ' < ./hex > ./hex2

#Adresse von dem Sensor in HEX
echo Adresse Sensor .........: 0x$(cut -d' ' -f1 hex2)

# Seriennummer zum prüfen ob es der richtige Sensor ist
SN1=$(cut -d' ' -f6 hex2)
SN2=$(cut -d' ' -f5 hex2)
SN3=$(cut -d' ' -f12 hex2)
SN4=$(cut -d' ' -f11 hex2)
echo Seriennummer ...........: $(( 16#$SN1$SN2$SN3$SN4 ))

# Calibration Year
YEAR1=$(cut -d' ' -f18 hex2)
YEAR2=$(cut -d' ' -f17 hex2)
echo Calibration Year .......: $(( 16#$YEAR1$YEAR2 ))

# Calibration  Mounth
MONTH=$(cut -d' ' -f24 hex2)
echo Calibration Month ......: $(( 16#$MONTH ))

# Calibration  Day
DAY=$(cut -d' ' -f23 hex2)
echo Calibration Day ........: $(( 16#$DAY ))

# Upper Range of pressure URP
URP1=$(cut -d' ' -f30 hex2)  # Achtung hier ist die Reihenfolge anders 
URP2=$(cut -d' ' -f29 hex2)  # noch anzupassen
URP3=$(cut -d' ' -f36 hex2)
URP4=$(cut -d' ' -f35 hex2)
URP=$(echo -ne "\x"$URP4"\x"$URP3"\x"$URP2"\x"$URP1 | hexdump -e '1/4 "%f" "\n"')
echo Upper Range of pressure : $URP

# Lower Range of pressure LRP
LRP1=$(cut -d' ' -f42 hex2)  # Achtung hier ist die Reihenfolge anders 
LRP2=$(cut -d' ' -f41 hex2)  # noch anzupassen
LRP3=$(cut -d' ' -f48 hex2)
LRP4=$(cut -d' ' -f47 hex2)
LRP=$(echo -ne "\x"$LRP4"\x"$LRP3"\x"$LRP2"\x"$LRP1 | hexdump -e '1/4 "%f" "\n"')
echo Lower Range of pressure : $LRP

# Actual pressure AP
AP1=$(cut -d' ' -f54 hex2)  # Achtung hier ist die Reihenfolge anders 
AP2=$(cut -d' ' -f53 hex2)  # noch anzupassen
AP3=$(cut -d' ' -f60 hex2)
AP4=$(cut -d' ' -f59 hex2)
#echo $AP4 $AP3 $AP2 $AP1
AP=$(echo -ne "\x"$AP4"\x"$AP3"\x"$AP2"\x"$AP1 | hexdump -e '1/4 "%f" "\n"')
echo Actal pressure .........: $AP

# Maximal pressure MAP
MAP1=$(cut -d' ' -f66 hex2)  # Achtung hier ist die Reihenfolge anders 
MAP2=$(cut -d' ' -f65 hex2)  # noch anzupassen
MAP3=$(cut -d' ' -f72 hex2)
MAP4=$(cut -d' ' -f71 hex2)
#echo $MAP4 $MAP3 $MAP2 $MAP1
MAP=$(echo -ne "\x"$MAP4"\x"$MAP3"\x"$MAP2"\x"$MAP1 | hexdump -e '1/4 "%f" "\n"')
echo Maximal pressure .......: $MAP

# Minimal pressure MIP
MIP1=$(cut -d' ' -f78 hex2)  # Achtung hier ist die Reihenfolge anders 
MIP2=$(cut -d' ' -f77 hex2)  # noch anzupassen
MIP3=$(cut -d' ' -f84 hex2)
MIP4=$(cut -d' ' -f83 hex2)
#echo $MIP4 $MIP3 $MIP2 $MIP1
MIP=$(echo -ne "\x"$MIP4"\x"$MIP3"\x"$MIP2"\x"$MIP1 | hexdump -e '1/4 "%f" "\n"')
echo Minimal pressure ......: $MIP
