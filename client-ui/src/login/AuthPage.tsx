import * as React from "react";
import {useNavigate, useSearchParams} from "react-router-dom";
import {useGetTokenMutation} from "../redux/api.js";
import {setCredentials} from "../redux/auth.js"
import {useEffect} from "react";
import {useDispatch} from "react-redux";

type GetTokenRequest = {
    email: string,
    uuid: string,
}

export function AuthPage() {
    const [searchParams, setSearchParams] = useSearchParams()

    const email = searchParams.get('email')
    const uuid = searchParams.get('uuid')
    const returnPath = searchParams.get("return")

    console.log(email, uuid, returnPath)

    const [getToken] = useGetTokenMutation()

    const navigate = useNavigate();
    const dispatch = useDispatch();

    useEffect(() => {
        // Your code here
        const body: GetTokenRequest = {
            email: email,
            uuid: uuid,
        }

        getToken(body).then((res: any) => {
            if (res?.error) {
                console.log("error", res)
            } else {
                const token = res?.data?.token
                console.log("success:", token)
                dispatch(
                    setCredentials({token})
                )
                navigate(returnPath ? returnPath : "/", {replace: true})
            }
        })
    }, []);

    return <div/>
}