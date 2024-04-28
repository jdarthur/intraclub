import {FlagOutlined, HomeOutlined, TrophyOutlined} from "@ant-design/icons";
import {Link, useLocation} from "react-router-dom";
import * as React from "react";
import {Menu} from "antd";
import {ROOT, LEAGUE, TEAM} from "./router"
import {useToken} from "../redux/auth.js";
import {UserIconSelfFetching} from "./NavBarUserIcon";

export function NavMenu() {

    const {pathname} = useLocation()

    const auth = useToken()
    return <Menu theme="dark" mode="horizontal" style={{flex: 1, minWidth: 0}} selectedKeys={[pathname]}>
        <Menu.Item key={ROOT} icon={<HomeOutlined/>}>
            <Link to={ROOT}>Home</Link>
        </Menu.Item>
        <Menu.Item key={LEAGUE} icon={<TrophyOutlined/>}>
            <Link to={LEAGUE}>Leagues</Link>
        </Menu.Item>
        <Menu.Item key={TEAM} icon={<FlagOutlined/>}>
            <Link to={TEAM}>Teams</Link>
        </Menu.Item>
        {auth ? <Menu.Item key="user">
            <UserIconSelfFetching/>
        </Menu.Item> : null}
    </Menu>
}