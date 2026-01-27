package fact

import (
	"sync"

	"github.com/iand/gdate"
	"github.com/iand/genster/model"
)

type NamedDay string

const (
	NamedDayNone           NamedDay = ""
	NamedDayEasterSunday   NamedDay = "easter sunday"
	NamedDayChristmasDay   NamedDay = "christmas day"
	NamedDayLadyDay        NamedDay = "lady day"
	NamedDayMichaelmas     NamedDay = "michaelmas"
	NamedDayCandlemas      NamedDay = "candlemas"
	NamedDayEpiphany       NamedDay = "epiphany"
	NamedDayAllSaintsDay   NamedDay = "all saints day"
	NamedDayMaundyThursday NamedDay = "maundy thursday"
	NamedDayPalmSunday     NamedDay = "palm sunday"
	NamedDayGoodFriday     NamedDay = "good friday"
	NamedDayAshWednesday   NamedDay = "ash wednesday"
	NamedDayAscensionDay   NamedDay = "ascension day"
	NamedDayWhitsunday     NamedDay = "whitsunday"
	NamedDayTrinitySunday  NamedDay = "trinity sunday"
)

func (n NamedDay) String() string {
	switch n {

	case NamedDayNone:
		return ""
	case NamedDayEasterSunday:
		return "Easter Sunday"
	case NamedDayChristmasDay:
		return "Christmas Day"
	case NamedDayLadyDay:
		return "Lady Day"
	case NamedDayMichaelmas:
		return "Michaelmas"
	case NamedDayCandlemas:
		return "Candlemas"
	case NamedDayEpiphany:
		return "Epiphany"
	case NamedDayAllSaintsDay:
		return "All Saints Day"
	case NamedDayMaundyThursday:
		return "Maundy Thursday"
	case NamedDayPalmSunday:
		return "Palm Sunday"
	case NamedDayGoodFriday:
		return "Good Friday"
	case NamedDayAshWednesday:
		return "Ash Wednesday"
	case NamedDayAscensionDay:
		return "Ascension Day"
	case NamedDayWhitsunday:
		return "Whitsun"
	case NamedDayTrinitySunday:
		return "Trinity Sunday"
	default:
		panic("unsupported named day value: " + n)
	}
}

func LookupNamedDay(dt *model.Date) NamedDay {
	p, ok := dt.Date.(*gdate.Precise)
	if !ok {
		return NamedDayNone
	}

	return lookupPrecise(p)
}

func lookupPrecise(p *gdate.Precise) NamedDay {
	initDateLookupsOncer.Do(initDateLookups)

	jd := p.C.JulianDay(p.Y, p.M, p.D)
	if named, ok := dayLookups[jd]; ok {
		return named
	}

	return DateLookups[jd]
}

// daynum calculates a quick unique number for each day of the year
func daynum(m int, d int) int {
	return m*31 + d
}

var dayLookups = map[int]NamedDay{
	daynum(1, 6):   NamedDayEpiphany,
	daynum(2, 2):   NamedDayCandlemas,
	daynum(3, 25):  NamedDayLadyDay,
	daynum(9, 29):  NamedDayMichaelmas,
	daynum(11, 1):  NamedDayAllSaintsDay,
	daynum(12, 25): NamedDayChristmasDay,
}

// DateLookups is a nested map of julian day to a NamedDay
var DateLookups = map[int]NamedDay{}

var initDateLookupsOncer sync.Once

func initDateLookups() {
	// dates of easter sunday
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1550, M: 4, D: 6})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1551, M: 3, D: 29})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1552, M: 4, D: 17})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1553, M: 4, D: 2})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1554, M: 3, D: 25})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1555, M: 4, D: 14})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1556, M: 4, D: 5})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1557, M: 4, D: 18})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1558, M: 4, D: 10})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1559, M: 3, D: 26})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1560, M: 4, D: 14})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1561, M: 4, D: 6})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1562, M: 3, D: 29})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1563, M: 4, D: 11})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1564, M: 4, D: 2})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1565, M: 4, D: 22})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1566, M: 4, D: 14})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1567, M: 3, D: 30})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1568, M: 4, D: 18})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1569, M: 4, D: 10})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1570, M: 3, D: 26})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1571, M: 4, D: 15})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1572, M: 4, D: 6})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1573, M: 3, D: 22})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1574, M: 4, D: 11})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1575, M: 4, D: 3})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1576, M: 4, D: 22})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1577, M: 4, D: 7})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1578, M: 3, D: 30})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1579, M: 4, D: 19})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1580, M: 4, D: 3})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1581, M: 3, D: 26})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1582, M: 4, D: 15})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1583, M: 3, D: 31})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1584, M: 4, D: 19})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1585, M: 4, D: 11})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1586, M: 4, D: 3})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1587, M: 4, D: 16})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1588, M: 4, D: 7})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1589, M: 3, D: 30})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1590, M: 4, D: 19})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1591, M: 4, D: 4})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1592, M: 3, D: 26})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1593, M: 4, D: 15})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1594, M: 3, D: 31})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1595, M: 4, D: 20})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1596, M: 4, D: 11})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1597, M: 3, D: 27})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1598, M: 4, D: 16})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1599, M: 4, D: 8})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1600, M: 3, D: 23})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1601, M: 4, D: 12})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1602, M: 4, D: 4})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1603, M: 4, D: 24})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1604, M: 4, D: 8})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1605, M: 3, D: 31})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1606, M: 4, D: 20})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1607, M: 4, D: 5})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1608, M: 3, D: 27})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1609, M: 4, D: 16})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1610, M: 4, D: 8})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1611, M: 3, D: 24})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1612, M: 4, D: 12})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1613, M: 4, D: 4})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1614, M: 4, D: 24})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1615, M: 4, D: 9})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1616, M: 3, D: 31})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1617, M: 4, D: 20})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1618, M: 4, D: 5})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1619, M: 3, D: 28})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1620, M: 4, D: 16})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1621, M: 4, D: 1})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1622, M: 4, D: 21})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1623, M: 4, D: 13})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1624, M: 3, D: 28})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1625, M: 4, D: 17})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1626, M: 4, D: 9})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1627, M: 3, D: 25})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1628, M: 4, D: 13})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1629, M: 4, D: 5})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1630, M: 3, D: 28})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1631, M: 4, D: 10})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1632, M: 4, D: 1})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1633, M: 4, D: 21})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1634, M: 4, D: 6})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1635, M: 3, D: 29})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1636, M: 4, D: 17})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1637, M: 4, D: 9})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1638, M: 3, D: 25})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1639, M: 4, D: 14})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1640, M: 4, D: 5})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1641, M: 4, D: 25})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1642, M: 4, D: 10})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1643, M: 4, D: 2})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1644, M: 4, D: 21})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1645, M: 4, D: 6})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1646, M: 3, D: 29})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1647, M: 4, D: 18})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1648, M: 4, D: 2})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1649, M: 3, D: 25})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1650, M: 4, D: 14})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1651, M: 3, D: 30})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1652, M: 4, D: 18})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1653, M: 4, D: 10})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1654, M: 3, D: 26})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1655, M: 4, D: 15})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1656, M: 4, D: 6})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1657, M: 3, D: 29})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1658, M: 4, D: 11})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1659, M: 4, D: 3})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1660, M: 4, D: 22})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1661, M: 4, D: 14})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1662, M: 3, D: 30})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1663, M: 4, D: 19})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1664, M: 4, D: 10})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1665, M: 3, D: 26})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1666, M: 4, D: 15})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1667, M: 4, D: 7})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1668, M: 3, D: 22})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1669, M: 4, D: 11})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1670, M: 4, D: 3})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1671, M: 4, D: 23})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1672, M: 4, D: 7})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1673, M: 3, D: 30})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1674, M: 4, D: 19})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1675, M: 4, D: 4})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1676, M: 3, D: 26})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1677, M: 4, D: 15})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1678, M: 3, D: 31})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1679, M: 4, D: 20})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1680, M: 4, D: 11})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1681, M: 4, D: 3})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1682, M: 4, D: 16})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1683, M: 4, D: 8})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1684, M: 3, D: 30})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1685, M: 4, D: 19})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1686, M: 4, D: 4})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1687, M: 3, D: 27})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1688, M: 4, D: 15})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1689, M: 3, D: 31})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1690, M: 4, D: 20})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1691, M: 4, D: 12})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1692, M: 3, D: 27})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1693, M: 4, D: 16})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1694, M: 4, D: 8})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1695, M: 3, D: 24})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1696, M: 4, D: 12})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1697, M: 4, D: 4})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1698, M: 4, D: 24})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1699, M: 4, D: 9})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1700, M: 3, D: 31})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1701, M: 4, D: 20})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1702, M: 4, D: 5})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1703, M: 3, D: 28})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1704, M: 4, D: 16})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1705, M: 4, D: 8})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1706, M: 3, D: 24})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1707, M: 4, D: 13})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1708, M: 4, D: 4})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1709, M: 4, D: 24})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1710, M: 4, D: 9})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1711, M: 4, D: 1})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1712, M: 4, D: 20})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1713, M: 4, D: 5})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1714, M: 3, D: 28})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1715, M: 4, D: 17})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1716, M: 4, D: 1})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1717, M: 4, D: 21})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1718, M: 4, D: 13})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1719, M: 3, D: 29})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1720, M: 4, D: 17})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1721, M: 4, D: 9})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1722, M: 3, D: 25})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1723, M: 4, D: 14})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1724, M: 4, D: 5})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1725, M: 3, D: 28})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1726, M: 4, D: 10})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1727, M: 4, D: 2})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1728, M: 4, D: 21})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1729, M: 4, D: 6})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1730, M: 3, D: 29})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1731, M: 4, D: 18})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1732, M: 4, D: 9})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1733, M: 3, D: 25})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1734, M: 4, D: 14})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1735, M: 4, D: 6})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1736, M: 4, D: 25})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1737, M: 4, D: 10})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1738, M: 4, D: 2})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1739, M: 4, D: 22})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1740, M: 4, D: 6})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1741, M: 3, D: 29})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1742, M: 4, D: 18})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1743, M: 4, D: 3})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1744, M: 3, D: 25})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1745, M: 4, D: 14})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1746, M: 3, D: 30})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1747, M: 4, D: 19})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1748, M: 4, D: 10})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1749, M: 3, D: 26})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1750, M: 4, D: 15})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1751, M: 4, D: 7})
	addMoveableFeasts(&gdate.Precise{C: gdate.Julian25Mar, Y: 1752, M: 3, D: 29})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1753, M: 4, D: 22})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1754, M: 4, D: 14})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1755, M: 3, D: 30})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1756, M: 4, D: 18})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1757, M: 4, D: 10})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1758, M: 3, D: 26})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1759, M: 4, D: 15})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1760, M: 4, D: 6})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1761, M: 3, D: 22})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1762, M: 4, D: 11})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1763, M: 4, D: 3})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1764, M: 4, D: 22})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1765, M: 4, D: 7})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1766, M: 3, D: 30})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1767, M: 4, D: 19})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1768, M: 4, D: 3})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1769, M: 3, D: 26})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1770, M: 4, D: 15})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1771, M: 3, D: 31})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1772, M: 4, D: 19})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1773, M: 4, D: 11})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1774, M: 4, D: 3})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1775, M: 4, D: 16})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1776, M: 4, D: 7})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1777, M: 3, D: 30})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1778, M: 4, D: 19})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1779, M: 4, D: 4})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1780, M: 3, D: 26})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1781, M: 4, D: 15})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1782, M: 3, D: 31})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1783, M: 4, D: 20})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1784, M: 4, D: 11})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1785, M: 3, D: 27})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1786, M: 4, D: 16})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1787, M: 4, D: 8})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1788, M: 3, D: 23})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1789, M: 4, D: 12})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1790, M: 4, D: 4})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1791, M: 4, D: 24})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1792, M: 4, D: 8})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1793, M: 3, D: 31})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1794, M: 4, D: 20})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1795, M: 4, D: 5})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1796, M: 3, D: 27})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1797, M: 4, D: 16})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1798, M: 4, D: 8})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1799, M: 3, D: 24})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1800, M: 4, D: 13})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1801, M: 4, D: 5})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1802, M: 4, D: 18})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1803, M: 4, D: 10})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1804, M: 4, D: 1})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1805, M: 4, D: 14})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1806, M: 4, D: 6})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1807, M: 3, D: 29})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1808, M: 4, D: 17})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1809, M: 4, D: 2})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1810, M: 4, D: 22})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1811, M: 4, D: 14})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1812, M: 3, D: 29})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1813, M: 4, D: 18})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1814, M: 4, D: 10})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1815, M: 3, D: 26})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1816, M: 4, D: 14})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1817, M: 4, D: 6})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1818, M: 3, D: 22})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1819, M: 4, D: 11})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1820, M: 4, D: 2})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1821, M: 4, D: 22})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1822, M: 4, D: 7})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1823, M: 3, D: 30})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1824, M: 4, D: 18})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1825, M: 4, D: 3})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1826, M: 3, D: 26})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1827, M: 4, D: 15})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1828, M: 4, D: 6})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1829, M: 4, D: 19})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1830, M: 4, D: 11})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1831, M: 4, D: 3})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1832, M: 4, D: 22})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1833, M: 4, D: 7})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1834, M: 3, D: 30})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1835, M: 4, D: 19})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1836, M: 4, D: 3})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1837, M: 3, D: 26})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1838, M: 4, D: 15})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1839, M: 3, D: 31})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1840, M: 4, D: 19})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1841, M: 4, D: 11})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1842, M: 3, D: 27})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1843, M: 4, D: 16})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1844, M: 4, D: 7})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1845, M: 3, D: 23})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1846, M: 4, D: 12})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1847, M: 4, D: 4})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1848, M: 4, D: 23})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1849, M: 4, D: 8})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1850, M: 3, D: 31})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1851, M: 4, D: 20})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1852, M: 4, D: 11})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1853, M: 3, D: 27})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1854, M: 4, D: 16})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1855, M: 4, D: 8})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1856, M: 3, D: 23})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1857, M: 4, D: 12})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1858, M: 4, D: 4})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1859, M: 4, D: 24})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1860, M: 4, D: 8})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1861, M: 3, D: 31})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1862, M: 4, D: 20})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1863, M: 4, D: 5})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1864, M: 3, D: 27})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1865, M: 4, D: 16})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1866, M: 4, D: 1})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1867, M: 4, D: 21})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1868, M: 4, D: 12})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1869, M: 3, D: 28})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1870, M: 4, D: 17})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1871, M: 4, D: 9})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1872, M: 3, D: 31})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1873, M: 4, D: 13})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1874, M: 4, D: 5})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1875, M: 3, D: 28})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1876, M: 4, D: 16})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1877, M: 4, D: 1})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1878, M: 4, D: 21})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1879, M: 4, D: 13})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1880, M: 3, D: 28})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1881, M: 4, D: 17})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1882, M: 4, D: 9})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1883, M: 3, D: 25})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1884, M: 4, D: 13})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1885, M: 4, D: 5})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1886, M: 4, D: 25})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1887, M: 4, D: 10})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1888, M: 4, D: 1})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1889, M: 4, D: 21})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1890, M: 4, D: 6})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1891, M: 3, D: 29})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1892, M: 4, D: 17})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1893, M: 4, D: 2})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1894, M: 3, D: 25})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1895, M: 4, D: 14})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1896, M: 4, D: 5})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1897, M: 4, D: 18})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1898, M: 4, D: 10})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1899, M: 4, D: 2})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1900, M: 4, D: 15})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1901, M: 4, D: 7})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1902, M: 3, D: 30})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1903, M: 4, D: 12})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1904, M: 4, D: 3})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1905, M: 4, D: 23})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1906, M: 4, D: 15})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1907, M: 3, D: 31})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1908, M: 4, D: 19})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1909, M: 4, D: 11})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1910, M: 3, D: 27})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1911, M: 4, D: 16})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1912, M: 4, D: 7})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1913, M: 3, D: 23})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1914, M: 4, D: 12})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1915, M: 4, D: 4})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1916, M: 4, D: 23})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1917, M: 4, D: 8})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1918, M: 3, D: 31})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1919, M: 4, D: 20})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1920, M: 4, D: 4})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1921, M: 3, D: 27})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1922, M: 4, D: 16})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1923, M: 4, D: 1})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1924, M: 4, D: 20})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1925, M: 4, D: 12})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1926, M: 4, D: 4})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1927, M: 4, D: 17})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1928, M: 4, D: 8})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1929, M: 3, D: 31})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1930, M: 4, D: 20})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1931, M: 4, D: 5})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1932, M: 3, D: 27})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1933, M: 4, D: 16})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1934, M: 4, D: 1})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1935, M: 4, D: 21})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1936, M: 4, D: 12})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1937, M: 3, D: 28})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1938, M: 4, D: 17})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1939, M: 4, D: 9})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1940, M: 3, D: 24})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1941, M: 4, D: 13})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1942, M: 4, D: 5})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1943, M: 4, D: 25})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1944, M: 4, D: 9})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1945, M: 4, D: 1})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1946, M: 4, D: 21})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1947, M: 4, D: 6})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1948, M: 3, D: 28})
	addMoveableFeasts(&gdate.Precise{C: gdate.Gregorian, Y: 1949, M: 4, D: 17})
}

func addMoveableFeasts(easterSunday *gdate.Precise) {
	jd := easterSunday.C.JulianDay(easterSunday.Y, easterSunday.M, easterSunday.D)

	DateLookups[jd] = NamedDayEasterSunday
	DateLookups[jd-2] = NamedDayGoodFriday
	DateLookups[jd-3] = NamedDayMaundyThursday
	DateLookups[jd-7] = NamedDayPalmSunday
	DateLookups[jd-46] = NamedDayAshWednesday
	DateLookups[jd+39] = NamedDayAscensionDay
	DateLookups[jd+49] = NamedDayWhitsunday
	DateLookups[jd+56] = NamedDayTrinitySunday
}
