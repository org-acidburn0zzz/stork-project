#!/bin/bash
echo -n "Searching for Prometheus ..."
if ! which prometheus > /dev/null; then
   echo -e "\nPrometheus not found! Install? (y/n) \c"
   read
   if "$REPLY" = "y"; then
      sudo apt-get install prometheus
      echo "Prometheus has been already installed."
   fi

else
    echo " yes."
fi

echo -n "Searching for Grafana ..."
if ! which grafana-server > /dev/null; then
   echo -e "\nGrafana not found! Install? (y/n) \c"
   read
   if "$REPLY" = "y"; then
      sudo apt-get install grafana
      echo "Grafana has been already installed."
   fi
else
    echo " yes."
fi

echo -n "Searching for Kea-exporter ..."
if ! which kea-exporter > /dev/null; then
   echo -e "\nKea-exporter not found! Install? (y/n) \c"
   read
   if "$REPLY" = "y"; then
      sudo pip3 install kea-exporter
      echo "Kea-exporter has been already installed."
   fi
else
    echo " yes."
fi
