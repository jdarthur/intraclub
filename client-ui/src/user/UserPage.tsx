import {useParams} from "react-router-dom";
import {useGetUserByIdQuery} from "../redux/api.js";
import * as React from "react";
import {Card, Skeleton} from "antd";
import {NavigationBreadcrumb} from "../navigation/NavigationBreadcrumb";
import {NameAndUserIcon} from "./UserIcon";
import {TeamsForUser} from "./TeamsForUser";
import {LabeledValue} from "../common/LabeledValue";
import {User} from "../settings/Users";
import {ExperienceForUser} from "./ExperienceForUser";
import {LoadingOutlined} from "@ant-design/icons";
import {Login} from "../login/Login";

export const USER_PAGE_WIDTH = "min(100vw - 2em, 650px)"

export function UserPage() {
    const {id} = useParams()

    const {data, isFetching} = useGetUserByIdQuery(id)

    const user: User = {
        email: data?.resource?.email,
        first_name: data?.resource?.first_name,
        last_name: data?.resource?.last_name,
        id: id,
        skill_info: data?.resource?.skill_info
    }

    const name = isFetching ? <LoadingOutlined /> :  `${user.first_name} ${user.last_name}`

    const title = <NameAndUserIcon FirstName={user.first_name} LastName={user.last_name} Loading={isFetching} />

    return <div>
        <NavigationBreadcrumb items={["Users", name]}/>

        <Card title={title} style={{width: USER_PAGE_WIDTH}}>
            {isFetching ? <Skeleton loading={isFetching} paragraph={{rows: 1}} /> : <LabeledValue label={"Email"} value={user.email} vertical/> }
        </Card>

        <div style={{height: "1em"}}/>

        <ExperienceForUser UserId={id}/>

        <div style={{height: "1em"}}/>

        <TeamsForUser UserId={id}/>

    </div>
}
