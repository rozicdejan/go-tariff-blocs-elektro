# Tariff Zone Display and API - Home Assistant Add-on

This project can also be installed and run as a Home Assistant add-on, providing integration for tariff zone data within Home Assistant. The add-on fetches tariff zone data and makes it available to Home Assistant sensors and dashboard entities for visualization and automation.

## Installing the Home Assistant Add-on

### Step 1: Clone or Download the Add-on

Navigate to your Home Assistant configuration directory, then clone or copy the repository:

```bash
cd /path/to/homeassistant/config/
mkdir -p addons/local
cd addons/local
git clone https://github.com/rozicdejan/go-tariff-blocs-elektro
```
### Step 2: Configure the Add-on
1. Navigate to the Home Assistant web interface.
2. Go to Settings > Add-ons > Add-on Store.
3. Click on the three-dot menu in the top-right corner and select Repositories.
4. Add the local path to the add-on directory:
```bash
/config/addons/local/tariff-zone-display
```
5. The add-on should now appear in the list of local add-ons. Click on it.

### Step 3: Start the Add-on
1. Click on Install to install the add-on.
2. Once installed, configure any necessary settings if applicable (e.g., server port, if exposed as a configurable option).
3. Click Start to launch the add-on.
4. You can view the add-on logs to ensure it is running correctly.

### Step 4: Verify API Access
To verify that the add-on is running, open a web browser and navigate to:
```bash
http://<your-home-assistant-ip>:8080
```
This should display the circular tariff zone display.


You can also access the API endpoint:
```bash
http://<your-home-assistant-ip>:8080/api/tariff
```

## Integrating with Home Assistant Entities
You can use the REST sensor in Home Assistant to fetch data from the add-on.

### Example REST Sensor Configuration
Add the following to your configuration.yaml file:

```yaml
sensor:
  - platform: rest
    name: Tariff Zone
    resource: http://<your-home-assistant-ip>:8080/api/tariff
    value_template: "{{ value_json.zone }}"
    json_attributes:
      - label
      - remaining_block_time
    scan_interval: 60  # Optional, time in seconds between updates
```
### Displaying Data in the Dashboard
You can create an Entities card in the Lovelace UI to display the tariff zone data:
```yaml
type: entities
title: Tariff Zone Information
entities:
  - entity: sensor.tariff_zone
    name: Current Tariff Zone
    icon: mdi:power-plug-battery-outline
  - type: attribute
    entity: sensor.tariff_zone
    attribute: label
    name: Season
    icon: mdi:message-fast-outline
  - type: attribute
    entity: sensor.tariff_zone
    attribute: remaining_block_time
    name: Remaining Time
    icon: mdi:timer
```

### Managing the Add-on
## Starting/Stopping the Add-on
To start or stop the add-on, go to the Settings > Add-ons > Tariff Zone Display, then click on the appropriate action (e.g., Start, Stop, Restart).

## Viewing Logs
To view the logs for the add-on, click on Log within the add-on management screen.
