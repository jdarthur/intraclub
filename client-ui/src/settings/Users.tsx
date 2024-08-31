import * as React from "react";
import {Space, Table} from "antd";
import {useGetUsersQuery} from "../redux/api.js";
import {DeleteConfirm} from "../common/DeleteConfirm";
import {EditOutlined, LoginOutlined} from "@ant-design/icons";
import {UserImport} from "./UserImport";

const columns = [
    {
        title: "Action",
        key: "action",
        dataIndex: "",
        render: (_: any, user: User) => <UserAction user_id={user.user_id} first_name={user.first_name}
                                                    last_name={user.last_name} email={user.email} skill_info={user.skill_info}/>,
    },
    {
        title: "First Name",
        key: "first_name",
        dataIndex: "first_name",
    },
    {
        title: "Last Name",
        key: "last_name",
        dataIndex: "last_name",
    },
    {
        title: "Email",
        key: "email",
        dataIndex: "email",
    }
]

export function Users() {
    const {data} = useGetUsersQuery()
    return <div>
        <UserImport />
        <Table dataSource={data?.resource} columns={columns}/>;
    </div>
}

export type User = {
    user_id: string
    first_name: string
    last_name: string
    email: string
    skill_info: string[]
}

function UserAction({user_id, first_name, last_name, email, skill_info}: User) {

    const deleteSelf = () => {
        console.log(`Delete user ${user_id}`)
    }

    const edit = () => {
        const u: User = {first_name, last_name, email, user_id, skill_info}
        console.log("Edit user", u)
    }

    return <Space>
        <a href={`/user/${user_id}`} key={"view"}>View</a>
        <DeleteConfirm deleteFunction={deleteSelf} objectType={"user"} key={"delete"}/>
        <EditOutlined style={{cursor: "pointer"}} key={"edit"} onClick={edit}/>
    </Space>
}