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

// AuthPage is a redirecting page that does the following:
//  - parse the `email`, `uuid`, and `return` path (optional) from the query params
//  - use the `email` / `uuid` values to request a token from the API
//  - get the token out of the response and set it in the global redux store
//  - redirect the user either to the `return` path or the homepage if unspecified
export function AuthPage() {
    const [searchParams] = useSearchParams()

    const email = searchParams.get('email')
    const uuid = searchParams.get('uuid')
    const returnPath = searchParams.get("return")

    const [getToken] = useGetTokenMutation()

    // navigate will allow us to programmatically navigate to either
    // the return path or the home page of the app after authentication
    const navigate = useNavigate();

    // dispatch will allow us to hook into the redux store and save the
    // token from the API into a global state variable that we can retrieve
    // on other pages via useToken()
    const dispatch = useDispatch();

    // this will run when the component is first rendered
    useEffect(() => {

        const body: GetTokenRequest = {
            email: email,
            uuid: uuid,
        }

        // call API
        getToken(body).then((res: any) => {
            if (res?.error) {
                // should probably show some kind of alert here and
                // have a link back to the login page
                console.log("error", res)
            } else {
                // parse token out of the response
                const token = res?.data?.token

                // set the token in the redux store
                dispatch(setCredentials({token}))

                // return back to the desired page
                navigate(returnPath ? returnPath : "/", {replace: true})
            }
        })
    }, []);

    return <div/>
}