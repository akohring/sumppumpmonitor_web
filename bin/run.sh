#!/usr/bin/env bash

# https://pinout.xyz
# Using physical pin notation

# TODO Post to remote server and monitor

EMAIL_ADDRESS=""
EMAIL_USER=""
SCRIPT_HOME=/opt/sumppumpmonitor
LOG_DIR=$SCRIPT_HOME/log
mkdir -p $LOG_DIR
LOG_FILE=$LOG_DIR/monitor_$(date +%Y.%m.%d).log
SUMP_PUMP_FAILED_FILE=$SCRIPT_HOME/sumppumpfailure

info() {
  log "[INFO] $@"
}
error() {
  log "[ERROR] $@"
}
log() {
  NOW=$(date +"%Y-%m-%d %H:%M:%S.%N")
  echo "$NOW $@" | tee -a $LOG_FILE
}
sendEmail() {
  NOW=$(date +"%Y-%m-%d %H:%M:%S.%N")
  SUBJECT="$NOW $1"
  TO="${EMAIL_ADDRESS}"
  BODY="$2"
  echo "${BODY}" | sudo -u "${EMAIL_USER}" mailx -s "${SUBJECT}" "${TO}"
  info "Email notification sent"
}

info "Start sump pump monitor"

find $LOG_DIR/*.log -type f -mtime +7 -exec rm -f {} \;

sumpPumpPin=11
gpio -1 mode $sumpPumpPin in

# Check sump pump pin
sumpPumpValue=$(gpio -1 read $sumpPumpPin)
alarmValue=0
if [ $sumpPumpValue = 0 ] ; then
  alarmValue=1
fi
if [ $alarmValue = 1 ] && [ ! -f $SUMP_PUMP_FAILED_FILE ]; then
  touch $SUMP_PUMP_FAILED_FILE
  sendEmail "Sump Pump Panic!" "Sump pump level is in the danger zone"
  error "Sump Pump Panic!"
elif [ $alarmValue = 1 ]; then
  error "Detected sump pump failure"
elif [ $alarmValue = 0 ] && [ -f $SUMP_PUMP_FAILED_FILE ]; then
  info "Sump Pump Recovered"
  rm -f $SUMP_PUMP_FAILED_FILE
  sendEmail "Sump Pump OK" "Sump pump level is safe"
else
  info "Sump Pump OK"
fi

RESP=$(curl -i -s --max-time 10 -X POST http://localhost:8080/pithealth/1/${sumpPumpValue})
EXITCODE=$?
if [ $EXITCODE -eq 0 ]; then
  info "Successfully updated sump pump health"
else
  error "Failed to post sump pump health; exitCode=${EXITCODE}; response=${RESP}"
fi
