
if [ $# -ne 2 ]; then
    die "Expected 2 argument (directory), not $#: $@"
fi

INPUT=${1}
OUTPUT=${2}

cat ${INPUT} | grep -v -E "sync progress skip|message is transaction|push transaction into TransactionPool from node|client|missing transactions|invalid voting basis|got ballot|stopped to handle ballot|module=network|BlockOperation|checked block|check block|stop checking; all the blocks are checked|check interval" > ${OUTPUT}

sed -i -E 's/ballot=[a-z0-9A-Z]+ //g' ${OUTPUT}

sed -i -E 's/module=[a-z0-9A-Z]+ //g' ${OUTPUT}

sed -i 's/GDIRF4UWPACXPPI4GW7CMTACTCNDIKJEHZK44RITZB4TD3YUM6CCVNGJ/n1/g' ${OUTPUT}
sed -i 's/GAYGELM74WJMKSLDN5YP2VAMP64WC4IXIGICUNK2SCVIT7KPTLY7M3MW/n2/g' ${OUTPUT}
sed -i 's/GDTEPFWEITKFHSUO44NQABY2XHRBBH2UBVGJ2ZJPDREIOL2F6RAEBJE4/n3/g' ${OUTPUT}
sed -i 's/GCDCXYUTLFOZSRQ4K6DTZ3R7ZJMD6CHVL3ZM7KGOG52NOTNHWBKYPCIO/n4/g' ${OUTPUT}
sed -i 's/GAUWM6QEXAS5JFPVC3SYGQKSW6S4SLCJEO7NX3JRDARHGJJBXIJYHTCX/n5/g' ${OUTPUT}

sed -i 's/GDIR.M6CC/n1/g' ${OUTPUT}
sed -i 's/GAYG.TLY7/n2/g' ${OUTPUT}
sed -i 's/GDTE.6RAE/n3/g' ${OUTPUT}
sed -i 's/GCDC.WBKY/n4/g' ${OUTPUT}
sed -i 's/GAUW.XIJY/n5/g' ${OUTPUT}