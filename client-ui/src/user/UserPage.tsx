import {useParams} from "react-router-dom";
import {useGetUserByIdQuery} from "../redux/api.js";
import * as React from "react";
import {Card, Divider} from "antd";
import {NavigationBreadcrumb} from "../navigation/NavigationBreadcrumb";
import {NameAndUserIcon} from "./UserIcon";
import {TeamsForUser} from "./TeamsForUser";

export function UserPage() {
    const {id} = useParams()

    const {data} = useGetUserByIdQuery(id)

    const firstName = data?.resource?.first_name
    const lastName = data?.resource?.last_name
    const email = data?.resource?.email

    const name = `${firstName} ${lastName}`

    const title = <NameAndUserIcon FirstName={firstName} LastName={lastName}/>

    return <div>
        <NavigationBreadcrumb items={["Users", name]}/>

        <Card title={title} style={{width: 300}}>
            {email}
        </Card>

        <div style={{height: "1em"}}/>

        <TeamsForUser UserId={id}/>
    </div>
}
