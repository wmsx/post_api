package setting

const XConfURL = "http://sx-xconf-micro:8090"
const DevXConfURL = "http://192.168.0.199:8090"

func SetUp(appName, env string) (err error) {
	if err = setUpMinio(appName, env); err != nil {
		return err
	}
	if err = setUpRedis(appName, env); err != nil {
		return err
	}
	return nil
}
