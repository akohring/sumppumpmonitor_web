#!/usr/bin/env python
import RPi.GPIO as GPIO
import requests
import time
from datetime import datetime

# Numbers GPIOs by physical location
GPIO.setmode(GPIO.BOARD)

# Define GPIO pins
TRIG_pin = 16
ECHO_pin = 18

# Setup GPIO I/O
GPIO.setup(TRIG_pin,GPIO.OUT)
GPIO.setup(ECHO_pin,GPIO.IN)
GPIO.setwarnings(False)

def distance( trig_mode ):
   # Wait for Sensor to Settle
   GPIO.output(TRIG_pin, False)
   time.sleep(2)

   # Set Trigger to High for 0.010ms
   GPIO.output(TRIG_pin, True)
   time.sleep(0.00001)
   GPIO.output(TRIG_pin, False)

   # Trig_mode = 1: Send 2nd trigger pulse with 0.05ms delay
   if trig_mode == 1:
      time.sleep(0.00005)
      GPIO.output(TRIG, True)
      time.sleep(0.00001)
      GPIO.output(TRIG, False)

   # Trig_mode = 2: Send 2nd trigger pulse with 0.075ms delay
   if trig_mode == 2:
      time.sleep(0.000075)
      GPIO.output(TRIG, True)
      time.sleep(0.00001)
      GPIO.output(TRIG, False)

   # Trig_mode = 3: Send 2nd trigger pulse with 0.1ms delay
   if trig_mode == 3:
      time.sleep(0.0001)
      GPIO.output(TRIG, True)
      time.sleep(0.00001)
      GPIO.output(TRIG, False)

   # Trig_mode = 4: Send 2nd trigger pulse with 0.125ms delay
   if trig_mode == 4:
      time.sleep(0.000125)
      GPIO.output(TRIG, True)
      time.sleep(0.00001)
      GPIO.output(TRIG, False)

   # Trig_mode = 5: Send 2nd trigger pulse with 0.15ms delay
   if trig_mode == 5:
      time.sleep(0.00015)
      GPIO.output(TRIG, True)
      time.sleep(0.00001)
      GPIO.output(TRIG, False)

   # Initialize pulse timers
   pulse_start = time.time()
   pulse_end = time.time()

   # Capture start of echo pulse
   while GPIO.input(ECHO_pin)==0:
      pulse_start = time.time()

   # Capture end of echo pulse
   while GPIO.input(ECHO_pin)==1:
      pulse_end = time.time()

   # Calculate time difference
   pulse_duration = pulse_end - pulse_start

   # Multiply time difference * speed of sound (1/2 out & back)
   distance = pulse_duration * 34300/2

   return distance

trig_mode = 0
dist_cm = -distance( trig_mode )
dist_in = dist_cm/2.54

for i in range(1,5):
   if dist_in<-19.5 or dist_in>-1:
      trig_mode = i
      dist_cm = -distance( trig_mode )
      dist_in = dist_cm/2.54

dist_cm = round(dist_cm,2)
dist_in = round(dist_in,2)

#with open("/opt/sumppumpmonitor/sumppumpdata.txt", "a") as f:
#	f.write(datetime.utcnow().strftime('%Y-%m-%d %H:%M:%S.%f') + " " + str(dist_in) + "\n")
requests.post("http://localhost:8080/pitlevel/1/"+str(dist_in))

GPIO.cleanup()
