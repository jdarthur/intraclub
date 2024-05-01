import * as React from 'react';
import {Button} from "antd";
import {LoginOutlined} from "@ant-design/icons";
import {Link} from "react-router-dom";

export function NavBarLogin() {
    return <Link to={"/login"}>
        <Button type={"primary"}>
            <LoginOutlined/>
            Login
        </Button>
    </Link>

}