#!/usr/bin/env bash
: "${DRIVE_FILES_DIR:="/home/vagrant/zfs/drive_files"}"
: "${POOL_NAME:="zsm_test"}"
: "${POOL_TYPE:="raidz2"}"
: "${POOL_SIZE:=4}"

TEST_DATA_SETS=(
    "$POOL_NAME/ds_1"
    "$POOL_NAME/ds_2"
    "$POOL_NAME/ds_2/nested_ds_1"
)

mkdir -p "$DRIVE_FILES_DIR"

# Create files to use as fake drives for zfs
declare -a DRIVE_FILES
for (( i=1; i<=POOL_SIZE; i++ )); do
    echo "Creating drive file $i/$POOL_SIZE"
    drive_file="$DRIVE_FILES_DIR/file_$i"
    dd if=/dev/zero of="$drive_file" bs=1G count=4 > /dev/null 2>&1
    DRIVE_FILES+=("$drive_file")
done

echo "Creating zfs pool from drive files"
zpool create "$POOL_NAME" "$POOL_TYPE" "${DRIVE_FILES[@]}"

# Create test data sets
for ds in "${TEST_DATA_SETS[@]}"; do
    echo "Creating test data set: $ds"
    zfs create "$ds"
done
