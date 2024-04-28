import {createApi, fetchBaseQuery} from '@reduxjs/toolkit/query/react';

export const baseQuery = fetchBaseQuery({
    baseUrl: '/api/',
    prepareHeaders: (headers, {getState}) => {
        const userToken = getState().auth.token
        if (userToken) {
            headers.set('x-session-token', userToken);
        }

        return headers;
    }
});

export const mainApi = createApi({
    reducerPath: 'mainApi',
    baseQuery: baseQuery,
    endpoints: (builder) => ({
        createOneTimePassword: builder.mutation({
            query: (body) => ({url: `one_time_password`, method: 'POST', body: body}),
        }),
        getToken: builder.mutation({
            query: (body) => ({
                url: 'token',
                method: 'POST',
                body: body,
            })
        }),
        whoAmI: builder.query({
            query: () => ({url: 'whoami'}),
            providesTags: ['whoami']
        }),
        getUsers: builder.query({
            query: () => ({url: `users`}),
            providesTags: ['users']
        }),
        getUserById: builder.query({
            query: (userId) => ({url: `users/${userId}`}),
        }),
        getTeamsByUserId: builder.query({
            query: (userId) => ({url: `teams_for_user/${userId}`}),
        }),
        getLeaguesByUserId: builder.query({
            query: (userId) => ({url: `leagues_for_user/${userId}`}),
        }),
        getLeaguesCommissionedByUserId: builder.query({
            query: (userId) => ({url: `leagues_commissioned_by_user/${userId}`}),
        })
    })
});

export const {
    useGetUsersQuery,
    useCreateOneTimePasswordMutation,
    useGetTokenMutation,
    useWhoAmIQuery,
    useGetUserByIdQuery,
    useGetTeamsByUserIdQuery,
    useGetLeaguesByUserIdQuery,
    useGetLeaguesCommissionedByUserIdQuery
} = mainApi
