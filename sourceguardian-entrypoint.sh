cd /xircl
git reset --hard
git pull
php bin/remove-files-with-annotoaions.php
echo $(find src -type f -exec md5sum {} \; | awk '{print $1}' | sort | md5sum | sed 's/ //g') > checksum
FILE=$(cat checksum)-php.zip
cd /root
curl -u $NEXTCLOUD_USER:$NEXTCLOUD_PASSWORD \
    $NEXTCLOUD_HOST/remote.php/dav/files/$NEXTCLOUD_USER/php-sourceguardian/$FILE \
    --output build.zip
if [ $(find build.zip -printf "%s") -gt 300 ] ;then
    echo "$FILE exists. skipping..."
else
    echo "$FILE does not exist."
    rm build.zip
    rm -rf /build
    sourceguardian/sourceguardian -r -o /build -j "<?php echo \"Copyright by SANDSIV SWITZERLAND LTD. Unknown error. Pleae contact with administrator\"; exit(1); ?>" --docker-socket /var/run/sourceguardian-docker.sock --phpversion 7.4 /xircl/src
    zip -r $FILE /build/xircl/src
    curl -u $NEXTCLOUD_USER:$NEXTCLOUD_PASSWORD -T $FILE $NEXTCLOUD_HOST/remote.php/dav/files/$NEXTCLOUD_USER/php-sourceguardian/$FILE
fi
