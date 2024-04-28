import {UserOutlined} from "@ant-design/icons";
import {Avatar, Popover} from "antd";
import * as React from "react";
import {ProduceCssColorFromHashedName} from "../navigation/NavBarUserIcon";

type User = {
    FirstName: string
    LastName: string
    Email: string
    UserId: string
}

export function UserIcon({FirstName, LastName, Email}: User) {
    const color = ProduceCssColorFromHashedName(FirstName, LastName)

    const title = <NameAndUserIcon FirstName={FirstName} LastName={LastName}/>

    const content = <div>{Email}</div>

    return <Popover title={title} content={content}>
        <Avatar style={{backgroundColor: color}} size={"large"}>
            {FirstName[0]}{LastName[0]}
        </Avatar>
    </Popover>
}

type FirstNameLastName = {
    FirstName: string,
    LastName: string,
}

export function NameAndUserIcon({FirstName, LastName}: FirstNameLastName) {
    return <span>
        <UserOutlined style={{marginRight: "1em"}}/>
        {FirstName} {LastName}
    </span>
}