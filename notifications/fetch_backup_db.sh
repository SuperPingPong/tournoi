#!/bin/bash

BACKUP_DIR=$PWD
rsync -avz scaleway:/root/backup $BACKUP_DIR
