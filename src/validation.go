package main

func IsValidProfileName(name string) bool {
    for _, profile := range profiles {
        if profile.Name == name {
            return true
        }
    }
    return false
}

func IsValidRegionName(name string) bool {

    for _, region := range RegionsMock {
        if region == name {
            return true
        }
    }

    return false
}