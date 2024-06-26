## **Croned Twicasting Recorder mp4**

Checks the live status of streamers on twitcasting.tv automatically at scheduled time, and records the live stream if
it's available  
If ffmpeg installed, converts the .ts file to .mp4 file.

---

### **Disclaimer**

This application constantly calls unofficial, non-documented twitcasting API to fetch live stream status. Please note
that:

* This application might not work in the future, subjecting to any change of twitcasting APIs
* Checking live stream status at high frequency might result in being banned from using twitcasting service, subjecting
  to twitcasting's terms and condition

<span style="color:red">Please note the above and use this application at your own risk. </span>

---

### [Docker support](https://hub.docker.com/r/kenken0803kr/croned-twitcasting-recorder-mp4/tags)
Docker image with ffmpeg pre-installed is available.
* It is recommended to mount the `/tw/file` path, which is the file storage path inside the container.
* It is recommended to using the environment variable TZ to set it to the same time zone as your host.
* If you want to use the direct mode, you need to specify `"direct"` argument and `"-streamer="`  [Usage](#usage)
* If you want to use the croned mode, you need to mount the `/tw/config.yaml`

```Bash
docker pull kenken0803kr/croned-twitcasting-recorder-mp4:latest

# direct mode
docker run -v ./:/tw/file kenken0803kr/croned-twitcasting-recorder-mp4 direct -streamer=azusa_shirokyan

# croned mode (require config.yaml)
docker run -v ./:/tw/file -v ./config.yaml:/tw/config.yaml kenken0803kr/croned-twitcasting-recorder-mp4
```


### docker-compose
```yaml
version: '3'
services:
  croned-twitcasting-recorder:
    image: kenken0803kr/croned-twitcasting-recorder-mp4:latest
    environment:
      - TZ=Asia/Seoul
    volumes:
      - ./:/tw/file
      - ./config.yaml:/tw/config.yaml

```

---

### **Requirements**

* **ffmpeg Installation**   
  ffmpeg should be installed and added to the system's PATH for each platform:
    - **Linux:** Install ffmpeg by running the following command in your terminal:
      ```
      sudo apt install ffmpeg
      ```

    - **macOS:** If you use Homebrew, you can install ffmpeg with the following command:
      ```
      brew install ffmpeg
      ```

    - **Windows:** If you use Windows 11, winget is already installed.
      ```
      winget install --id=Gyan.FFmpeg  -e
      ```

---

### **Installation**

* **Executables**   
  Executables can be found on [release page](https://github.com/KENKEN0803/croned-twitcasting-recorder-mp4/releases).
* **Build from source**   
  Ensure that [golang is installed](https://golang.org/doc/install) on your system.
  ```Bash
  git clone https://github.com/jzhang046/croned-twitcasting-recorder && cd croned-twitcasting-recorder
  go build -o ./bin/
  # Executable: ./bin/croned-twitcasting-recorder
  ```

--- 

### **Usage**

**Croned recording mode _(default)_**  
* Make sure config.yaml is in the same path as the executable file.
* Please refer to [configuration](#configuration) section below to create configuration file.
  ```Bash
  # Grant execution permission
  chmod 755 ./bin/croned-twitcasting-recorder-mp4
  
  # Execute below command to start the recorder
  ./bin/croned-twitcasting-recorder-mp4

  # Or specify croned recording mode explicitly 
  ./bin/croned-twitcasting-recorder-mp4 croned
  ```
* In Windows, simply place the .exe file and config.yaml file in the same path and double-click the exe file.


**Direct recording mode**  
  Direct recording mode supports recording to start immediately, with configurable number of retries and retry backoff
  period.
  ```Bash
  # Start in direct recording mode  
  ./bin/croned-twitcasting-recorder direct --streamer=${STREAMER_SCREEN_ID}
  """
  Usage of direct:
  -retries int
    	[optional] number of retries (default 0)
  -retry-backoff duration
    	[optional] retry backoff period (default 15s)
  -streamer string
    	[required] streamer URL
  -encode-option string
          [optional] ffmpeg video encode option. (default copy)
  """
  # Streamer URL must be supplied as argument 

  # Example: 
  ./bin/croned-twitcasting-recorder direct --streamer=azusa_shirokyan --retries=10 --retry-backoff=1m --encode-option="libx265 -preset ultrafast"
  ```

---

### **Configuration**

Configuration file `config.yaml` must be present on the current directory under croned recording mode. Please
see [config_example.yaml] for example format.  
At least 1 streamer should be specified in `config.yaml`  
Multiple streamers could be specified with individual schedules. Status check and recording for different streamers
would _not_ affect each other.

#### Field explanations:

+ `screen-id`:  
  Presented on the URL of the screamer's top page.  
  Example: Top page URL of streamer [小野寺梓@真っ白なキャンバス](https://twitcasting.tv/azusa_shirokyan)
  is `https://twitcasting.tv/azusa_shirokyan`, the corresponding screen-id is `azusa_shirokyan`
+ `schedule`:   
  Please refer to the below docs for supported schedule definitions:
    - https://pkg.go.dev/github.com/robfig/cron/v3#hdr-CRON_Expression_Format
    - https://pkg.go.dev/github.com/robfig/cron/v3#hdr-Predefined_schedules
+ `encode-option`:  
  If not provided, copy the stream without encoding and rebuild the .ts file to mp4 file.  
  See full documentation at
    - https://ffmpeg.org/ffmpeg-codecs.html#toc-Video-Encoders
    - https://trac.ffmpeg.org/wiki/Encode/H.265

---

### **Output**

Output recording file would be put under the ./file/{screen-id}/ directory, named
after `{StreamTitle}-yyyyMMdd-HHmm.ts`  
For example, a recording starts at 15:04 on 2nd Jan 2006 of
streamer [小野寺梓@真っ白なキャンバス](https://twitcasting.tv/azusa_shirokyan) would create recording
file `./file/azusa_shirokyan/{StreamTitle}-20060102-1504.ts`  
If ffmpeg is installed, .mp4 file is created instead of .ts file of the same name.
