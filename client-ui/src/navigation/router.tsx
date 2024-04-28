import {createBrowserRouter} from "react-router-dom";
import {MainLayout} from "./layout.js";
import App from "../App.js";
import {Login} from "../login/Login.js";
import * as React from "react";
import {AuthPage} from "../login/AuthPage.js";
import {UserPage} from "../user/UserPage";
import {AllLeaguesPage} from "../league/AllLeaguesPage";
import {TeamsPage} from "../team/TeamsPage";

export const ROOT = "/"
export const LOGIN = "/login"
export const LEAGUE = "/league"
export const TEAM = "/team"
export const AUTH = "/auth"
export const USER = "/user/:id"


export const router = createBrowserRouter([
    {
        path: ROOT,
        element: <MainLayout content={<App/>}/>
    },
    {
        path: LOGIN,
        element: <MainLayout content={<Login/>}/>
    },
    {
        path: LEAGUE,
        element: <MainLayout content={<AllLeaguesPage/>}/>
    },
    {
        path: TEAM,
        element: <MainLayout content={<TeamsPage/>}/>
    },
    {
        path: AUTH,
        element: <MainLayout content={<AuthPage/>}/>
    },
    {
        path: USER,
        element: <MainLayout content={<UserPage/>}/>
    }

]);