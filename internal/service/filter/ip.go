package filter

import (
	"context"
	"fmt"
	"net"
	"time"
)

func QueryPtr(ip string) (string, error) {
	ctx := context.Background()
	key := fmt.Sprintf("ip:ptr:%s", ip)

	// get from redis first
	if v, err := rdb.Get(ctx, key).Result(); err == nil {
		return v, nil
	}

	// get from dns
	data, err := resolver.LookupPtr(ip)
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", nil
	}

	// then save to redis
	rdb.Set(ctx, key, data[0], 24*time.Hour)
	return data[0], nil
}

type keyValue struct {
	key   string
	value any
}

/*
QueryRegion
---------------------
Record sample:

	city:
	  geoname_id: 1809858
	  names:
	    de: Guangzhou
	    en: Guangzhou
	    es: Cantón
	    fr: Canton
	    ja: 広州
	    pt-BR: Cantão
	    ru: Гуанчжоу
	    zh-CN: 广州市
	continent:
	  code: AS
	  geoname_id: 6255147
	  names:
	    de: Asien
	    en: Asia
	    es: Asia
	    fr: Asie
	    ja: アジア
	    pt-BR: Ásia
	    ru: Азия
	    zh-CN: 亚洲
	country:
	  geoname_id: 1814991
	  iso_code: CN
	  names:
	    de: China
	    en: China
	    es: China
	    fr: Chine
	    ja: 中国
	    pt-BR: China
	    ru: Китай
	    zh-CN: 中国
	location:
	  accuracy_radius: 100
	  latitude: 23.1181
	  longitude: 113.2539
	  time_zone: Asia/Shanghai
	registered_country:
	  geoname_id: 1814991
	  iso_code: CN
	  names:
	    de: China
	    en: China
	    es: China
	    fr: Chine
	    ja: 中国
	    pt-BR: China
	    ru: Китай
	    zh-CN: 中国
	subdivisions: [map[geoname_id:1809935 iso_code:GD names:map[en:Guangdong fr:Province de Guangdong zh-CN:广东]]]
*/
func QueryRegion(ip net.IP) (country, province, city string, err error) {
	ctx := context.Background()
	keyCountry := fmt.Sprintf("%s:%s:%s", ip, "region", "country")
	keyProvince := fmt.Sprintf("%s:%s:%s", ip, "region", "province")
	keyCity := fmt.Sprintf("%s:%s:%s", ip, "region", "city")

	// get from redis first
	if country, err = rdb.Get(ctx, keyCountry).Result(); err == nil {
		if province, err = rdb.Get(ctx, keyProvince).Result(); err == nil {
			if city, err = rdb.Get(ctx, keyCity).Result(); err == nil {
				return country, province, city, nil
			}
		}
	}

	// Get data of IP.
	anyData := make(map[string]any)
	_, ok, err := geoip.LookupNetwork(ip, &anyData)
	if err != nil {
		return "", "", "", err
	}

	for k, v := range anyData {
		if k == "country" {
			if vv, ok := v.(map[string]any); ok {
				if vvv, ok := vv["names"].(map[string]any); ok {
					if vvvv, ok := vvv["en"].(string); ok {
						country = vvvv
					}
				}
			}
		} else if k == "subdivisions" {
			if len(v.([]any)) > 0 {
				for _, sd := range v.([]any) {
					for kk, vv := range sd.(map[string]any) {
						if kk == "names" {
							if vvv, ok := vv.(map[string]any); ok {
								if vvvv, ok := vvv["en"].(string); ok {
									province = vvvv
								}
							}
						}
					}
				}
			}
		} else if k == "city" {
			if vv, ok := v.(map[string]any); ok {
				if vvv, ok := vv["names"].(map[string]any); ok {
					if vvvv, ok := vvv["en"].(string); ok {
						city = vvvv
					}
				}
			}
		}
	}
	if ok {
		return country, province, city, nil
	}

	return "", "", "", fmt.Errorf("not found")

}
