import service from '@/utils/request'

export const createRidgeContact = (data) => {
  return service({
    url: '/ridgeContact/createRidgeContact',
    method: 'post',
    data
  })
}

export const deleteRidgeContact = (params) => {
  return service({
    url: '/ridgeContact/deleteRidgeContact',
    method: 'delete',
    params
  })
}

export const deleteRidgeContactByIds = (params) => {
  return service({
    url: '/ridgeContact/deleteRidgeContactByIds',
    method: 'delete',
    params
  })
}

export const updateRidgeContact = (data) => {
  return service({
    url: '/ridgeContact/updateRidgeContact',
    method: 'put',
    data
  })
}

export const findRidgeContact = (params) => {
  return service({
    url: '/ridgeContact/findRidgeContact',
    method: 'get',
    params
  })
}

export const getRidgeContactList = (params) => {
  return service({
    url: '/ridgeContact/getRidgeContactList',
    method: 'get',
    params
  })
}

export const getRidgeContactStat = () => {
  return service({
    url: '/ridgeContact/getRidgeContactStat',
    method: 'get',
    donNotShowLoading: true
  })
}

export const getUpcomingContactBirthdays = (params) =>
  service({
    url: '/ridgeContact/getUpcomingContactBirthdays',
    method: 'get',
    params,
    donNotShowLoading: true
  })
