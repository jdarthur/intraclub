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
        }),
        getFacilities: builder.query({
            query: () => ({url: `facilities`}),
            providesTags: ["facilities"]
        }),
        createFacility: builder.mutation({
            query: (body) => ({url: `facilities`, body: body, method: 'POST'}),
            invalidatesTags: ["facilities"]
        }),
        updateFacility: builder.mutation({
            query: (req) => ({url: `facilities/${req.id}`, body: req.body, method: 'put'}),
            invalidatesTags: ["facilities"]
        }),
        deleteFacility: builder.mutation({
            query: (id) => ({url: `facilities/${id}`, method: 'DELETE'}),
            invalidatesTags: ["facilities"]
        }),
        createWeek: builder.mutation({
            query: (body) => ({url: `weeks`, method: 'POST', body: body}),
            invalidatesTags: ["weeks"]
        }),
        createLeague: builder.mutation({
            query: (body) => ({url: `leagues`, method: 'POST', body: body}),
            invalidatesTags: ["leagues"]
        }),
        getWeekById: builder.query({
            query: (id) => ({url: `weeks/${id}`}),
        }),
        getFacilityById: builder.query({
            query: (id) => ({url: `facilities/${id}`}),
        }),

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
    useGetLeaguesCommissionedByUserIdQuery,
    useGetFacilitiesQuery,
    useCreateFacilityMutation,
    useDeleteFacilityMutation,
    useUpdateFacilityMutation,
    useCreateWeekMutation,
    useCreateLeagueMutation,
    useGetWeekByIdQuery,
    useGetFacilityByIdQuery,
} = mainApi
