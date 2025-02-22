import * as React from "react"
import {NavigationBreadcrumb} from "../navigation/NavigationBreadcrumb";
import {Facilities} from "./Facilities";
import {useToken} from "../redux/auth.js";
import {LoginRequired} from "../login/LoginRequired";
import {Tabs} from "antd";
import {Users} from "./Users";


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
            children: <Users />,
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