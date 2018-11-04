package main

import(
  "os"
  "bufio"
  "strconv"
  "time"
  "os/exec"
  "strings"
)

func getBatteryPercentage() int {
  file, err := os.Open("/sys/class/power_supply/BAT0/capacity")
  if err != nil {
    return -1
  }

  defer file.Close()
  c := bufio.NewScanner(file)
  c.Scan()

  percentage, err := strconv.Atoi(c.Text())
  if err != nil {
    return -1
  } else {
    return percentage
  }
}

func getPowerStatus() bool {
  file, err := os.Open("/sys/class/power_supply/BAT0/status")
  if err != nil {
    return false
  }

  defer file.Close()
  c := bufio.NewScanner(file)
  c.Scan()

  switch status := c.Text(); status {
  case "Charging", "Full", "Unknown":
    return true
  default:
    return false
  }
}

func sendNotification(notification string) bool {
  args := []string{"-a", "powermon", "-u"}
  switch notification {
  case "Discharging":
    args = append(args, "normal", "Descarregando")
  case "Charging":
    args = append(args, "low", "Carregando")
  default:
    args = append(args, "critical", "Bateria Fraca")
  }

  batteryPercentageMsg := []string{"Bateria ", "em ", strconv.Itoa(getBatteryPercentage()), "%"}
  args = append(args, strings.Join(batteryPercentageMsg, ""))

  if err := exec.Command("notify-send", args...).Run(); err != nil {
    return false
  } else {
    return true
  }
}

func main() {
  type timerProfile struct {
    poolingIntervalSec int
    notifyIntervalMin int
  }

  type powerProfile struct {
    notify bool
    batteryPercentage int
    charger bool
  }

  timer := timerProfile {poolingIntervalSec: 2, notifyIntervalMin: 5}
  power := powerProfile {notify: false, batteryPercentage: 15, charger: false}
  notifyCounter := 0

  for {
    if getPowerStatus() {
      notifyCounter = 0
      if !power.charger {
        sendNotification("Charging")
        power.charger = true
        power.notify = true
      }
    } else {
      if power.charger {
        sendNotification("Discharging")
        power.charger = false
        power.notify = false
      }

      if getBatteryPercentage() <= power.batteryPercentage && !power.notify {
        sendNotification("*")
        power.notify = true
        notifyCounter = 0
      }

      if (notifyCounter * timer.poolingIntervalSec) >= (timer.notifyIntervalMin * 60) {
        sendNotification("*")
        notifyCounter = 0
      }
    }

    notifyCounter++
    time.Sleep(time.Duration(timer.poolingIntervalSec) * time.Second)
  }
}
