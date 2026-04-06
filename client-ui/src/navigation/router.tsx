import {createBrowserRouter} from "react-router-dom";
import {MainLayout} from "./layout.js";
import App from "../App.js";
import {Login} from "../login/Login.js";
import * as React from "react";
import {AuthPage} from "../login/AuthPage.js";
import {UserPage} from "../user/UserPage";
import {TeamsPage} from "../team/TeamsPage";
import {SettingsPage} from "../settings/SettingsPage";
import {Register} from "../login/Register";
import {DefaultScoreboard} from "../match_scoreboard/DefaultScoreboard";
import {Seasons} from "../season/Seasons";

export const ROOT = "/"
export const LOGIN = "/login"
export const REGISTER = "/register"
export const SEASONS = "/seasons"
export const TEAM = "/team"
export const AUTH = "/auth"
export const USER = "/user/:id"
export const SETTINGS = "/settings"
export const SCOREBOARD = "/scoreboard"
export const SEASON = "/season/:id"

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
        path: REGISTER,
        element: <MainLayout content={<Register/>}/>
    },
    {
        path: SEASONS,
        element: <MainLayout content={<Seasons />}/>
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
    },
    {
        path: SETTINGS,
        element: <MainLayout content={<SettingsPage/>}/>
    },
    {
        path: SCOREBOARD,
        element: <DefaultScoreboard/>
    },
    {
        path: SEASON,
        element: <MainLayout content={<>empty</>}/>
    }

]);