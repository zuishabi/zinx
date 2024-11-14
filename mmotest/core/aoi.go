package core

// 定义一些AOI的边界值
const (
	AOI_MIN_X  int = 85
	AOI_MAX_X  int = 410
	AOI_CNTS_X int = 10
	AOI_MIN_Y  int = 75
	AOI_MAX_Y  int = 400
	AOI_CNTS_Y int = 20
)

// AOI区域管理模块
type AOIManager struct {
	//区域的左边界坐标
	MinX int
	//区域右边界坐标
	MaxX int
	//X方向格子的个数
	CntsX int
	//区域的上边界坐标
	MinY int
	//区域的下边界坐标
	MaxY int
	//Y方向格子的个数
	CntsY int
	//当前区域中有哪些格子map-key=格子的ID，value=格子对象
	grids map[int]*Grid
}

// 创建一个AOI管理模块
func NewAOIManager(minX int, maxX int, cntsX int, minY int, maxY int, cntsY int) *AOIManager {
	aoiMgr := &AOIManager{
		MinX:  minX,
		MaxX:  maxX,
		CntsX: cntsX,
		MinY:  minY,
		MaxY:  maxY,
		CntsY: cntsY,
		grids: make(map[int]*Grid),
	}
	//给当前aoi初始化区域的格子所有的格子进行编号和初始化
	for y := 0; y < cntsY; y++ {
		for x := 0; x < cntsX; x++ {
			//计算格子ID，根据x，y编号
			gid := y*cntsX + x
			//初始化gid的格子
			aoiMgr.grids[gid] = NewGrid(gid, aoiMgr.MinX+x*aoiMgr.gridWidth(), aoiMgr.MinX+(x+1)*aoiMgr.gridWidth(),
				aoiMgr.MinY+y*aoiMgr.gridLength(), aoiMgr.MinY+(y+1)*aoiMgr.gridLength(),
			)
		}
	}
	return aoiMgr
}

// 得到每个格子在X轴方向的宽度
func (m *AOIManager) gridWidth() int {
	return (m.MaxX - m.MinX) / m.CntsX
}

// 得到每个格子在Y轴方向的长度
func (m *AOIManager) gridLength() int {
	return (m.MaxY - m.MinY) / m.CntsY
}

// 根据格子gid得到周边九宫格格子的集合
func (m *AOIManager) GetSurroundingGridsByGid(gID int) (grids []*Grid) {
	//判断当前gID是否在AOIManager中
	if _, ok := m.grids[gID]; !ok {
		return
	}
	//初始化grids返回值切片，将自身加入到九宫格切片中
	grids = append(grids, m.grids[gID])
	//需要gid的左边是否还有格子,右边是否有格子
	//需要通过gID得到当前格子x轴的编号，--idx = ID%nx
	idx := gID % m.CntsX
	//判断idx编号是否左边还有格子
	if idx > 0 {
		grids = append(grids, m.grids[gID-1])
	}
	//判断idx编号是否右边还有格子
	if idx < m.CntsX-1 {
		grids = append(grids, m.grids[gID+1])
	}
	//将x轴当前集合的值全部取出，进行遍历，再判断上下是否还有格子
	//先将x轴当前格子的ID集合
	gidsX := make([]int, 0, len(grids))
	for _, v := range grids {
		gidsX = append(gidsX, v.GID)
	}
	//遍历gidsX集合中每个格子的gid
	for _, v := range gidsX {
		//得到当前格子的id的y轴的编号 idy = id/ny
		idy := v / m.CntsY
		//gid上边是否还有格子
		if idy > 0 {
			grids = append(grids, m.grids[v-m.CntsX])
		}
		//gid下边是否还有格子
		if idy < m.CntsY-1 {
			grids = append(grids, m.grids[v+m.CntsX])
		}
	}
	return
}

// 根据横纵坐标得到当前GID格子编号
func (m *AOIManager) GetGidByPos(x, y float32) int {
	idx := (int(x) - m.MinX) / m.gridWidth()
	idy := (int(y) - m.MinY) / m.gridLength()
	return idx + idy
}

// 通过横纵坐标得到周边九宫格内全部的PlayerIDs
func (m *AOIManager) GetPidsByPos(x, y float32) (playerIDs []int) {
	//得到当前玩家的格子id
	gid := m.GetGidByPos(x, y)
	//通过gid得到周边九宫格信息
	grids := m.GetSurroundingGridsByGid(gid)
	//将九宫格的信息里的全部的player的id加到playerids中
	for _, v := range grids {
		playerIDs = append(playerIDs, v.GetPlayerIDs()...) //这里打三个点代表和playerIDs进行拼接
	}
	return
}

// 添加一个PlayerID到一个格子中
func (m *AOIManager) AddPidToGrid(pID, gID int) {
	m.grids[gID].Add(pID)
}

// 移除一个格子中的PlayerID
func (m *AOIManager) RemovePidFromGrid(pID, gID int) {
	m.grids[gID].Remove(pID)
}

// 通过GID获取全部的PlayerID
func (m *AOIManager) GetPidsByGid(gID int) (playerIDs []int) {
	playerIDs = m.grids[gID].GetPlayerIDs()
	return
}

// 通过坐标将Player添加到一个格子中
func (m *AOIManager) AddPidToGridByPos(pID int, x, y float32) {
	gID := m.GetGidByPos(x, y)
	grid := m.grids[gID]
	grid.Add(pID)
}

// 通过坐标把一个Player从格子中删除
func (m *AOIManager) RemovePidFromGridByPos(pID int, x, y float32) {
	gID := m.GetGidByPos(x, y)
	grid := m.grids[gID]
	grid.Remove(pID)
}
