import * as React from "react"
import {NavigationBreadcrumb} from "../navigation/NavigationBreadcrumb";
import {Facilities} from "./Facilities";
import {useToken} from "../redux/auth";
import {LoginRequired} from "../login/LoginRequired";

export function SettingsPage() {
    const token = useToken()

    return <div>
        <NavigationBreadcrumb items={["Settings"]}/>
        {token ? <Facilities/> : <LoginRequired/>}
    </div>

}