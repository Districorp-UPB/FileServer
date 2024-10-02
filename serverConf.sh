#! /bin/bash
sudo apt-get update
sudo apt install -y nfs-kernel-server nfs-common
sudo systemctl start nfs-kernel-server.service
sudo mkdir -p ./nfs/files
sudo mount ip_servidornfs:/ruta_nfs/carpeta ./nfs/files

