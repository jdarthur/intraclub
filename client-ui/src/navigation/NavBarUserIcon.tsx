import {Link} from "react-router-dom";
import * as React from "react";
import {logoutUser, useToken} from "../redux/auth.js";
import {useWhoAmIQuery} from "../redux/api.js";
import {UserIcon} from "../user/UserIcon";
import {Button} from "antd";
import {LogoutOutlined, SettingOutlined} from "@ant-design/icons";
import {useDispatch} from "react-redux";


export function UserIconSelfFetching() {
    const token = useToken()
    const {data} = useWhoAmIQuery({}, {skip: !token})

    const first_name = data?.first_name
    const last_name = data?.last_name
    const user_id = data?.user_id
    const email = data?.email


    const show = token && data

    return <div>
        {show ? <UserIcon UserId={user_id} Email={email} FirstName={first_name}
                          LastName={last_name} UseLink ShowLogout/> : null}
    </div>

}


// ProduceCssColorFromHashedName creates a color hex code e.g. #5042e6 from the provided
// first name and last name by doing a simple hash on the character codes in the name.
// This can be used to produce a stable color for a particular person's name e.g. in a
// user icon or player in a list.
export function ProduceCssColorFromHashedName(first_name: string, last_name: string): string {

    const firstNameLastName = `${first_name}${last_name}`;

    let hash = 0
    for (let i = 0; i < firstNameLastName.length; i++) {
        let char = firstNameLastName.charCodeAt(i);
        hash = ((hash << 5) - hash) + char;
        hash = hash & hash;
    }

    hash = Math.abs(hash)

    let colorCode = `#${hash.toString(16)}`
    if (colorCode.length > 7) {
        colorCode = colorCode.substring(0, 7)
    }


    return colorCode
}
