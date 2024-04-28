import * as React from 'react';
import {NavigationBreadcrumb} from "../navigation/NavigationBreadcrumb";
import {LeaguesForUser} from "./LeaguesForUser";
import {useWhoAmIQuery} from "../redux/api";
import {useToken} from "../redux/auth";
import {LoginRequired} from "../login/LoginRequired";
import {Button, Tabs, TabsProps} from "antd";
import {PlusOutlined, PlusSquareFilled, PlusSquareOutlined} from "@ant-design/icons";
import {LeaguesCommissionedByUser} from "./LeaguesCommissionedByUser";

export function AllLeaguesPage() {

    const token = useToken()
    const {data} = useWhoAmIQuery({}, {skip: !token})

    const items: TabsProps['items'] = [
        {
            key: "member_of",
            label: "Member",
            children: <LeaguesForUser UserId={data?.user_id}/>,
        },
        {
            key: "commissioned_by",
            label: "Commissioner",
            children: <LeaguesCommissionedByUser UserId={data?.user_id}/>,
        },
    ]

    const leaguesView = <Tabs defaultActiveKey="member_of" items={items}/>

    return <div>
        <NavigationBreadcrumb items={["My Leagues", "All"]}/>
        {token ? leaguesView : <LoginRequired/>}
    </div>
}