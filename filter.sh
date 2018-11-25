cat 1-1.txt | grep -v -E "missing transactions|invalid voting basis|got ballot|stopped to handle ballot|module=network|BlockOperation|checked block|check block|stop checking; all the blocks are checked|check interval" > 1-1-1.txt

cat 1-1-1.txt | grep -E "block hash different!!!!" > error.txt

cat 1-1-1.txt | grep "node=GDIR.M6CC" > 1-2-1.txt
cat 1-1-1.txt | grep "node=GAYG.TLY7" > 1-2-2.txt
cat 1-1-1.txt | grep "node=GDTE.6RAE" > 1-2-3.txt
cat 1-1-1.txt | grep "node=GCDC.WBKY" > 1-2-4.txt

sed -i -E 's/ballot=[a-z0-9A-Z]+ //g' 1-2-1.txt
sed -i -E 's/ballot=[a-z0-9A-Z]+ //g' 1-2-2.txt
sed -i -E 's/ballot=[a-z0-9A-Z]+ //g' 1-2-3.txt
sed -i -E 's/ballot=[a-z0-9A-Z]+ //g' 1-2-4.txt

sed -i -E 's/module=[a-z0-9A-Z]+ //g' 1-2-1.txt
sed -i -E 's/module=[a-z0-9A-Z]+ //g' 1-2-2.txt
sed -i -E 's/module=[a-z0-9A-Z]+ //g' 1-2-3.txt
sed -i -E 's/module=[a-z0-9A-Z]+ //g' 1-2-4.txt

sed -i "s/node=GDIR.M6CC //g" 1-2-1.txt
sed -i "s/node=GAYG.TLY7 //g" 1-2-2.txt
sed -i "s/node=GDTE.6RAE //g" 1-2-3.txt
sed -i "s/node=GCDC.WBKY //g" 1-2-4.txt

sed -i 's/GDIRF4UWPACXPPI4GW7CMTACTCNDIKJEHZK44RITZB4TD3YUM6CCVNGJ/n1/g' 1-2-1.txt
sed -i 's/GAYGELM74WJMKSLDN5YP2VAMP64WC4IXIGICUNK2SCVIT7KPTLY7M3MW/n2/g' 1-2-1.txt
sed -i 's/GDTEPFWEITKFHSUO44NQABY2XHRBBH2UBVGJ2ZJPDREIOL2F6RAEBJE4/n3/g' 1-2-1.txt
sed -i 's/GCDCXYUTLFOZSRQ4K6DTZ3R7ZJMD6CHVL3ZM7KGOG52NOTNHWBKYPCIO/n4/g' 1-2-1.txt

sed -i 's/GDIRF4UWPACXPPI4GW7CMTACTCNDIKJEHZK44RITZB4TD3YUM6CCVNGJ/n1/g' 1-2-2.txt
sed -i 's/GAYGELM74WJMKSLDN5YP2VAMP64WC4IXIGICUNK2SCVIT7KPTLY7M3MW/n2/g' 1-2-2.txt
sed -i 's/GDTEPFWEITKFHSUO44NQABY2XHRBBH2UBVGJ2ZJPDREIOL2F6RAEBJE4/n3/g' 1-2-2.txt
sed -i 's/GCDCXYUTLFOZSRQ4K6DTZ3R7ZJMD6CHVL3ZM7KGOG52NOTNHWBKYPCIO/n4/g' 1-2-2.txt

sed -i 's/GDIRF4UWPACXPPI4GW7CMTACTCNDIKJEHZK44RITZB4TD3YUM6CCVNGJ/n1/g' 1-2-3.txt
sed -i 's/GAYGELM74WJMKSLDN5YP2VAMP64WC4IXIGICUNK2SCVIT7KPTLY7M3MW/n2/g' 1-2-3.txt
sed -i 's/GDTEPFWEITKFHSUO44NQABY2XHRBBH2UBVGJ2ZJPDREIOL2F6RAEBJE4/n3/g' 1-2-3.txt
sed -i 's/GCDCXYUTLFOZSRQ4K6DTZ3R7ZJMD6CHVL3ZM7KGOG52NOTNHWBKYPCIO/n4/g' 1-2-3.txt

sed -i 's/GDIRF4UWPACXPPI4GW7CMTACTCNDIKJEHZK44RITZB4TD3YUM6CCVNGJ/n1/g' 1-2-4.txt
sed -i 's/GAYGELM74WJMKSLDN5YP2VAMP64WC4IXIGICUNK2SCVIT7KPTLY7M3MW/n2/g' 1-2-4.txt
sed -i 's/GDTEPFWEITKFHSUO44NQABY2XHRBBH2UBVGJ2ZJPDREIOL2F6RAEBJE4/n3/g' 1-2-4.txt
sed -i 's/GCDCXYUTLFOZSRQ4K6DTZ3R7ZJMD6CHVL3ZM7KGOG52NOTNHWBKYPCIO/n4/g' 1-2-4.txt

sed -i 's/GDIR.M6CC/n1/g' 1-2-1.txt
sed -i 's/GAYG.TLY7/n2/g' 1-2-1.txt
sed -i 's/GDTE.6RAE/n3/g' 1-2-1.txt
sed -i 's/GCDC.WBKY/n4/g' 1-2-1.txt

sed -i 's/GDIR.M6CC/n1/g' 1-2-2.txt
sed -i 's/GAYG.TLY7/n2/g' 1-2-2.txt
sed -i 's/GDTE.6RAE/n3/g' 1-2-2.txt
sed -i 's/GCDC.WBKY/n4/g' 1-2-2.txt

sed -i 's/GDIR.M6CC/n1/g' 1-2-3.txt
sed -i 's/GAYG.TLY7/n2/g' 1-2-3.txt
sed -i 's/GDTE.6RAE/n3/g' 1-2-3.txt
sed -i 's/GCDC.WBKY/n4/g' 1-2-3.txt

sed -i 's/GDIR.M6CC/n1/g' 1-2-4.txt
sed -i 's/GAYG.TLY7/n2/g' 1-2-4.txt
sed -i 's/GDTE.6RAE/n3/g' 1-2-4.txt
sed -i 's/GCDC.WBKY/n4/g' 1-2-4.txt