import {LogoutOutlined, UserOutlined} from "@ant-design/icons";
import {Avatar, Button, Divider, Popover} from "antd";
import * as React from "react";
import {ProduceCssColorFromHashedName} from "../navigation/NavBarUserIcon";
import {Link} from "react-router-dom";
import {logoutUser} from "../redux/auth";
import {useDispatch} from "react-redux";
import {c} from "vite/dist/node/types.d-aGj9QkWt";

type User = {
    FirstName: string
    LastName: string
    Email?: string
    UserId?: string
    UseLink?: boolean
    ShowLogout?: boolean
}

export function UserIcon({FirstName, LastName, Email, UserId, UseLink, ShowLogout}: User) {
    const color = ProduceCssColorFromHashedName(FirstName, LastName)

    const title = <NameAndUserIcon FirstName={FirstName} LastName={LastName} UserId={UserId} UseLink={UseLink}/>

    const logout = <Button type={"primary"} onClick={() => dispatch(logoutUser())}>
        <LogoutOutlined/>
        Log out
    </Button>

    const content = <div style={{display: "flex", flexDirection: "column"}}>
        <Divider style={{margin: '0px 0px 1em 0px'}}/>

        <div style={{marginBottom: "0.5em", textAlign: "center"}}>
            {Email}
        </div>
        {ShowLogout ? logout : null}
    </div>

    const dispatch = useDispatch()

    return <Popover title={title} content={content} style={{width: undefined}}>
        <Avatar style={{backgroundColor: color, flexGrow: 1}} size={"large"}>
            {FirstName[0]}{LastName[0]}
        </Avatar>

    </Popover>
}

export function NameAndUserIcon({FirstName, LastName, UserId, UseLink}: User) {

    const content = <span>
        <UserOutlined style={{marginRight: "1em"}}/>
        {FirstName} {LastName}
    </span>

    if (UseLink) {
        return <Link to={`/user/${UserId}`}>
            {content}
        </Link>
    }

    return content
}