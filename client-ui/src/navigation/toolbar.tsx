import {FlagOutlined, HomeOutlined, SettingOutlined, TrophyOutlined} from "@ant-design/icons";
import {Link, useLocation} from "react-router-dom";
import * as React from "react";
import {Menu} from "antd";
import {ROOT, LEAGUE, TEAM, SETTINGS, LOGIN} from "./router"
import {IntraclubTokenKey, setCredentials, useToken} from "../redux/auth.js";
import {UserIconSelfFetching} from "./NavBarUserIcon";
import {useEffect} from "react";
import {useDispatch} from "react-redux";
import {NavBarLogin} from "./NavBarLogin";

export function NavMenu() {
    const {pathname} = useLocation()

    const dispatch = useDispatch();

    useEffect(() => {
        const token = localStorage.getItem(IntraclubTokenKey)
        if (token != null) {
            dispatch(
                setCredentials({token})
            )
        }

    }, [])

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
        {auth ? <Menu.Item key={SETTINGS} icon={<SettingOutlined/>}>
            <Link to={SETTINGS}>Settings</Link>
        </Menu.Item> : null}
        <Menu.Item key="user">
            {auth ? <UserIconSelfFetching/> : pathname == LOGIN ? null : <NavBarLogin/>}
        </Menu.Item>
    </Menu>
}