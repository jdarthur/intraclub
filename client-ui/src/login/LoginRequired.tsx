import * as React from 'react';
import {Button, Card} from "antd";
import {Link, useLocation} from "react-router-dom";
import {LoginOutlined} from "@ant-design/icons";

export function LoginRequired() {

    const {pathname} = useLocation()

    return <Card size={"small"} title={"Login required"} style={{width: 200}} extra={<LoginOutlined/>}>
        <div style={{display: "flex", flexDirection: "column", alignItems: "flex-end"}}>
            <Link to={`/login?return=${pathname}`}>
                <Button type={"primary"} style={{alignSelf: "flex-end"}}>
                    Log in
                </Button>
            </Link>
        </div>
    </Card>
}