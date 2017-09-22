#!/usr/bin/env bash

package_name="watchcat"
platforms=("windows/amd64" "linux/amd64" "darwin/amd64")

for platform in "${platforms[@]}"
do
    pushd './cmd/watchcat'

    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    zip_name=$package_name'-'$GOOS'-'$GOARCH'.zip'
    output_name=$package_name
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi

    env GOOS=$GOOS GOARCH=$GOARCH go build -o '../../'$output_name
    if [ $? -ne 0 ]; then
        echo 'failed to build'
        exit 1
    fi

    popd

    chmod +x $output_name
    zip $zip_name $output_name
    if [ $? -ne 0 ]; then
        echo 'failed to zip'
        exit 1
    fi
    rm $output_name
    if [ $? -ne 0 ]; then
        echo 'failed to remove'
        exit 1
    fi
done
