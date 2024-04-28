import * as React from 'react';
import {Button, Card} from "antd";
import {Link, useLocation} from "react-router-dom";
import {Login} from "./Login";
import {LoginOutlined} from "@ant-design/icons";

export function LoginRequired() {

    const {pathname} = useLocation()

    return <Card size={"small"} title={"Login required"} style={{width: 200}} extra={<LoginOutlined/>}>
        <div style={{display: "flex", flexDirection: "column", width: "100%"}}>
            <Button type={"primary"} style={{alignSelf: "flex-end"}}>
                <Link to={`/login?return=${pathname}`}>
                    Log in
                </Link>
            </Button>
        </div>

    </Card>
}