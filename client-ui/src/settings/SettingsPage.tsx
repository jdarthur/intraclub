import * as React from "react"
import {NavigationBreadcrumb} from "../navigation/NavigationBreadcrumb";
import {Facilities} from "./facility/Facilities";
import {useToken} from "../redux/auth.js";
import {LoginRequired} from "../login/LoginRequired";
import {Tabs} from "antd";
import {Users} from "./Users";
import {ScoringStructures} from "./scoring_structure/ScoringStructures";
import {Ratings} from "./rating/Ratings";
import {Formats} from "./formats/Formats";


export function SettingsPage() {
    const token = useToken()

    const settingsTabs = [
        {
            label: "Facilities",
            key: "facilities",
            children: <Facilities/>,
        },
        {
            label: "Users",
            key: "users",
            children: <Users/>,
        },
        {
            label: "Scoring Structures",
            key: "scoring_structures",
            children: <ScoringStructures/>,
        },
        {
            label: "Ratings",
            key: "ratings",
            children: <Ratings/>,
        },
        {
            label: "Formats",
            key: "formats",
            children: <Formats/>,
        },
    ]

    const tabs = token ? <Tabs tabPosition={"left"}
                               style={{height: "100%"}}
                               tabBarStyle={{height: "100%"}}
                               items={settingsTabs}/> : <LoginRequired/>


    return <div style={{display: 'flex', flexDirection: "column", height: "100%"}}>
        <NavigationBreadcrumb items={["Settings"]}/>
        {tabs}
    </div>

}