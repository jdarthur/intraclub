import * as React from 'react';
import {useWhoAmIQuery} from "../redux/api";
import {NavigationBreadcrumb} from "../navigation/NavigationBreadcrumb";
import {useToken} from "../redux/auth";
import {TeamsForUser} from "../user/TeamsForUser";
import {LoginRequired} from "../login/LoginRequired";

export function TeamsPage() {

    const token = useToken()
    const {data} = useWhoAmIQuery({}, {skip: !token})

    return <div>
        <NavigationBreadcrumb items={["My Teams", "All"]}/>
        {token ? <TeamsForUser UserId={data?.user_id}/> : <LoginRequired/>}
    </div>
}