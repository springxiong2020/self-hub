export const formatContactBirthday = (contact = {}) => {
  if (!contact.birthday) return '未填写'
  if (!contact.birthdayIsLunar) return `公历 ${contact.birthday}`
  const [year, month, day] = contact.birthday.split('-').map(Number)
  const months = [
    '正',
    '二',
    '三',
    '四',
    '五',
    '六',
    '七',
    '八',
    '九',
    '十',
    '冬',
    '腊'
  ]
  const days = [
    '初一',
    '初二',
    '初三',
    '初四',
    '初五',
    '初六',
    '初七',
    '初八',
    '初九',
    '初十',
    '十一',
    '十二',
    '十三',
    '十四',
    '十五',
    '十六',
    '十七',
    '十八',
    '十九',
    '二十',
    '廿一',
    '廿二',
    '廿三',
    '廿四',
    '廿五',
    '廿六',
    '廿七',
    '廿八',
    '廿九',
    '三十'
  ]
  return `农历 ${year} 年${contact.birthdayIsLeapMonth ? '闰' : ''}${months[month - 1]}月${days[day - 1]}`
}
